package handler

import (
	"backend/internal/middleware"
	"backend/internal/repository"
	"backend/internal/service"
	"net/http"
	"sync"
	"time"
    "strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

// WSHandler 提供简单的基于用户ID的私信能力（内存转发，单实例）
type WSHandler struct {
	upgrader websocket.Upgrader
	mu       sync.RWMutex
	// 在线连接：userID -> set(conns)
	conns map[uuid.UUID]map[*websocket.Conn]struct{}
	jwtSvc service.JwtService
	friendRepo repository.FriendshipRepository
	roomRepo repository.ChatRoomRepository
	// 每个连接的写锁，避免并发写同一连接导致断开
	writeMu map[*websocket.Conn]*sync.Mutex
}

func NewWSHandler(jwtSvc service.JwtService, friendRepo repository.FriendshipRepository, roomRepo repository.ChatRoomRepository) *WSHandler {
	return &WSHandler{
		upgrader: websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			// 交由上游 Auth/CORS 控制，这里放宽跨域
			CheckOrigin: func(r *http.Request) bool { return true },
		},
		conns: make(map[uuid.UUID]map[*websocket.Conn]struct{}),
		jwtSvc: jwtSvc,
		friendRepo: friendRepo,
		roomRepo: roomRepo,
		writeMu: make(map[*websocket.Conn]*sync.Mutex),
	}
}

// inbound 消息结构
type inbound struct {
	ToUserID string `json:"to_user_id"`
	Content  string `json:"content"`
	RoomID   string `json:"room_id"`
}

// outbound 消息结构
type outbound struct {
	FromUserID string    `json:"from_user_id"`
	ToUserID   string    `json:"to_user_id"`
	Content    string    `json:"content"`
	Timestamp  time.Time `json:"timestamp"`
	RoomID     string    `json:"room_id"`
}

// Chat WebSocket 连接端点
//
// @Summary      WebSocket 聊天连接
// @Description  通过 WebSocket 建立一对一聊天连接。鉴权方式：优先读取 HTTP Header `Authorization: Bearer <access_token>`；浏览器无法自定义 Header 时可使用 query 参数 `?token=<access_token>`。
// @Tags         chat
// @Produce      json
// @Param        token   query   string  false  "Access Token（可选，浏览器场景使用）"
// @Success      101     {string}  string  "Switching Protocols"
// @Router       /ws/chat [get]
func (h *WSHandler) Chat(c *gin.Context) {
	var userID uuid.UUID
	if payload, ok := c.Get(middleware.AuthorizationPayloadKey); ok {
		if claims, ok := payload.(*service.JWTClaims); ok {
			userID = claims.UserID
		}
	}
	// 浏览器 WebSocket 无法自定义 Authorization 头，支持 query 参数 token 作为兜底
	if userID == uuid.Nil {
		// 兼容非浏览器客户端：优先从 Authorization: Bearer <token> 读取
		var token string
		if auth := c.GetHeader("Authorization"); auth != "" {
			lower := strings.ToLower(auth)
			if strings.HasPrefix(lower, "bearer ") && len(auth) > 7 {
				token = strings.TrimSpace(auth[7:])
			}
		}
		if token == "" {
			token = c.Query("token")
		}
		if token == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"message": "未授权"})
			return
		}
		claims, err := h.jwtSvc.ValidateToken(token)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"message": "Token 无效"})
			return
		}
		userID = claims.UserID
	}

	conn, err := h.upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}
	// 注册连接
	h.mu.Lock()
	set := h.conns[userID]
	if set == nil {
		set = make(map[*websocket.Conn]struct{})
		h.conns[userID] = set
	}
	set[conn] = struct{}{}
	h.writeMu[conn] = &sync.Mutex{}
	h.mu.Unlock()

	defer func() {
		h.mu.Lock()
		if s, ok := h.conns[userID]; ok {
			delete(s, conn)
			if len(s) == 0 { delete(h.conns, userID) }
		}
		delete(h.writeMu, conn)
		h.mu.Unlock()
		conn.Close()
	}()

	conn.SetReadLimit(64 * 1024)
	conn.SetReadDeadline(time.Now().Add(75 * time.Second))
	conn.SetPongHandler(func(string) error {
		conn.SetReadDeadline(time.Now().Add(75 * time.Second))
		return nil
	})

	// 心跳：服务端每30秒发送一次 ping，防止空闲超时与中间网络设备断开
	stopCh := make(chan struct{})
	go func(c *websocket.Conn) {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				h.mu.RLock()
				m := h.writeMu[c]
				h.mu.RUnlock()
				if m == nil { return }
				m.Lock()
				_ = c.SetWriteDeadline(time.Now().Add(10 * time.Second))
				err := c.WriteControl(websocket.PingMessage, []byte("ping"), time.Now().Add(10*time.Second))
				m.Unlock()
				if err != nil { return }
			case <-stopCh:
				return
			}
		}
	}(conn)

	// 读循环
	for {
		var msg inbound
		if err := conn.ReadJSON(&msg); err != nil {
			break
		}
		if msg.Content == "" { continue }

		var toID uuid.UUID
		var roomID uuid.UUID

		if msg.RoomID != "" {
			// 依据 room_id 发送，校验房间有效且自己为参与者
			rid, err := uuid.Parse(msg.RoomID)
			if err != nil { continue }
			room, err := h.roomRepo.GetByID(rid)
			if err != nil || room.Status != "active" { continue }
			if room.UserAID != userID && room.UserBID != userID { continue }
			if room.UserAID == userID { toID = room.UserBID } else { toID = room.UserAID }
			roomID = room.ID
		} else {
			// 依据 to_user_id 发送：先校验好友，再获取/创建房间
			var err error
			toID, err = uuid.Parse(msg.ToUserID)
			if err != nil { continue }
			// 需要存在好友关系
			if ok, err := h.friendRepo.Exists(userID, toID); err != nil || !ok {
				continue
			}
			room, err := h.roomRepo.GetOrCreateByUsers(userID, toID)
			if err != nil { continue }
			if room.Status != "active" { continue }
			roomID = room.ID
		}

		out := outbound{
			FromUserID: userID.String(),
			ToUserID:   toID.String(),
			Content:    msg.Content,
			Timestamp:  time.Now(),
			RoomID:     roomID.String(),
		}

		// 向目标用户与自己其他连接转发
		h.mu.RLock()
		recvSet := h.conns[toID]
		selfSet := h.conns[userID]
		h.mu.RUnlock()
		// 向对端写
		for ws := range recvSet {
			h.mu.RLock(); m := h.writeMu[ws]; h.mu.RUnlock()
			if m == nil { continue }
			m.Lock()
			_ = ws.SetWriteDeadline(time.Now().Add(10 * time.Second))
			_ = ws.WriteJSON(out)
			m.Unlock()
		}
		// 向自己其他连接写
		for ws := range selfSet {
			if ws == conn { continue }
			h.mu.RLock(); m := h.writeMu[ws]; h.mu.RUnlock()
			if m == nil { continue }
			m.Lock()
			_ = ws.SetWriteDeadline(time.Now().Add(10 * time.Second))
			_ = ws.WriteJSON(out)
			m.Unlock()
		}
	}
	close(stopCh)
}
