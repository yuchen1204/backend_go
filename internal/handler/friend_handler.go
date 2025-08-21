package handler

import (
	"backend/internal/middleware"
	"backend/internal/model"
	"backend/internal/response"
	"backend/internal/service"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// FriendHandler 好友系统处理器（MVP骨架）

type FriendHandler struct {
	friendSvc service.FriendService
}

func NewFriendHandler(friendSvc service.FriendService) *FriendHandler {
	return &FriendHandler{friendSvc: friendSvc}
}

// CreateRequest 发起好友请求
// @Summary 发起好友请求
// @Tags 好友
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Router /friends/requests [post]
func (h *FriendHandler) CreateRequest(c *gin.Context) {
	// 从token获取当前用户
	payload, ok := c.Get(middleware.AuthorizationPayloadKey)
	if !ok {
		response.ErrorResponse(c, http.StatusUnauthorized, "未授权", nil)
		return
	}
	claims, ok := payload.(*service.JWTClaims)
	if !ok {
		response.ErrorResponse(c, http.StatusUnauthorized, "授权信息错误", nil)
		return
	}

	// 解析请求
	var req struct {
		ReceiverID string `json:"receiver_id" binding:"required"`
		Note       string `json:"note"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "请求参数错误", err.Error())
		return
	}
	rid, err := uuid.Parse(req.ReceiverID)
	if err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "无效的接收方ID", err.Error())
		return
	}

	fr, err := h.friendSvc.CreateRequest(c.Request.Context(), claims.UserID, rid, req.Note)
	if err != nil {
		// 常见业务错误映射
		msg := err.Error()
		switch {
		case msg == "不能向自己发起请求" ||
			msg == "申请备注长度不能超过200字符" ||
			msg == "您当前被禁用发起好友申请" ||
			msg == "对方已将你拉黑，无法发起请求" ||
			msg == "请先取消对该用户的拉黑" ||
			msg == "您的好友数已达上限" ||
			msg == "对方的好友数已达上限" ||
			msg == "你们已经是好友" ||
			msg == "已存在待处理的请求" ||
			msg == "对方已有向你的待处理请求":
			response.ErrorResponse(c, http.StatusBadRequest, msg, nil)
		case msg == "请求过于频繁，请明天再试":
			response.ErrorResponse(c, http.StatusTooManyRequests, msg, nil)
		default:
			response.ErrorResponse(c, http.StatusInternalServerError, "创建好友请求失败", msg)
		}
		return
	}

	response.SuccessResponse(c, http.StatusCreated, "好友请求已发送", fr)
}

// AcceptRequest 接受好友请求
// @Summary 接受好友请求
// @Tags 好友
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param id path string true "请求ID"
// @Router /friends/requests/{id}/accept [post]
func (h *FriendHandler) AcceptRequest(c *gin.Context) {
	// 鉴权
	payload, ok := c.Get(middleware.AuthorizationPayloadKey)
	if !ok {
		response.ErrorResponse(c, http.StatusUnauthorized, "未授权", nil)
		return
	}
	claims, ok := payload.(*service.JWTClaims)
	if !ok {
		response.ErrorResponse(c, http.StatusUnauthorized, "授权信息错误", nil)
		return
	}

	// 路径参数
	idStr := c.Param("id")
	rid, ok := parseUUID(c, idStr)
	if !ok {
		return
	}

	if err := h.friendSvc.AcceptRequest(c.Request.Context(), rid, claims.UserID); err != nil {
		msg := err.Error()
		switch msg {
		case "请求不存在":
			response.ErrorResponse(c, http.StatusNotFound, msg, nil)
		case "请求已被处理", "无权处理该请求", "双方存在拉黑关系，无法成为好友", "对方的好友数已达上限", "您的好友数已达上限", "您当前被禁用发起好友申请":
			response.ErrorResponse(c, http.StatusBadRequest, msg, nil)
		default:
			response.ErrorResponse(c, http.StatusInternalServerError, "接受好友请求失败", msg)
		}
		return
	}
	response.SuccessResponse(c, http.StatusOK, "已接受好友请求", gin.H{"id": idStr})
}

// RejectRequest 拒绝好友请求
// @Summary 拒绝好友请求
// @Tags 好友
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param id path string true "请求ID"
// @Router /friends/requests/{id}/reject [post]
func (h *FriendHandler) RejectRequest(c *gin.Context) {
	payload, ok := c.Get(middleware.AuthorizationPayloadKey)
	if !ok {
		response.ErrorResponse(c, http.StatusUnauthorized, "未授权", nil)
		return
	}
	claims, ok := payload.(*service.JWTClaims)
	if !ok {
		response.ErrorResponse(c, http.StatusUnauthorized, "授权信息错误", nil)
		return
	}
	idStr := c.Param("id")
	rid, ok := parseUUID(c, idStr)
	if !ok {
		return
	}
	if err := h.friendSvc.RejectRequest(c.Request.Context(), rid, claims.UserID); err != nil {
		msg := err.Error()
		switch msg {
		case "请求不存在":
			response.ErrorResponse(c, http.StatusNotFound, msg, nil)
		case "请求已被处理", "无权处理该请求":
			response.ErrorResponse(c, http.StatusBadRequest, msg, nil)
		default:
			response.ErrorResponse(c, http.StatusInternalServerError, "拒绝好友请求失败", msg)
		}
		return
	}
	response.SuccessResponse(c, http.StatusOK, "已拒绝好友请求", gin.H{"id": idStr})
}

// CancelRequest 撤回好友请求
// @Summary 撤回好友请求
// @Tags 好友
// @Security ApiKeyAuth
// @Param id path string true "请求ID"
// @Router /friends/requests/{id} [delete]
func (h *FriendHandler) CancelRequest(c *gin.Context) {
	payload, ok := c.Get(middleware.AuthorizationPayloadKey)
	if !ok {
		response.ErrorResponse(c, http.StatusUnauthorized, "未授权", nil)
		return
	}
	claims, ok := payload.(*service.JWTClaims)
	if !ok {
		response.ErrorResponse(c, http.StatusUnauthorized, "授权信息错误", nil)
		return
	}
	idStr := c.Param("id")
	rid, ok := parseUUID(c, idStr)
	if !ok {
		return
	}
	if err := h.friendSvc.CancelRequest(c.Request.Context(), rid, claims.UserID); err != nil {
		msg := err.Error()
		switch msg {
		case "请求不存在":
			response.ErrorResponse(c, http.StatusNotFound, msg, nil)
		case "请求已被处理", "无权撤回该请求":
			response.ErrorResponse(c, http.StatusBadRequest, msg, nil)
		default:
			response.ErrorResponse(c, http.StatusInternalServerError, "撤回好友请求失败", msg)
		}
		return
	}
	response.SuccessResponse(c, http.StatusOK, "已撤回好友请求", gin.H{"id": idStr})
}

// ListFriends 好友列表
// @Summary 获取好友列表
// @Tags 好友
// @Security ApiKeyAuth
// @Produce json
// @Router /friends/list [get]
func (h *FriendHandler) ListFriends(c *gin.Context) {
	payload, ok := c.Get(middleware.AuthorizationPayloadKey)
	if !ok {
		response.ErrorResponse(c, http.StatusUnauthorized, "未授权", nil)
		return
	}
	claims, ok := payload.(*service.JWTClaims)
	if !ok {
		response.ErrorResponse(c, http.StatusUnauthorized, "授权信息错误", nil)
		return
	}
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	search := c.Query("search")
	list, total, err := h.friendSvc.ListFriends(c.Request.Context(), claims.UserID, search, page, limit)
	if err != nil {
		response.ErrorResponse(c, http.StatusInternalServerError, "获取好友列表失败", err.Error())
		return
	}
	data := map[string]any{
		"items": list,
		"total": total,
		"page":  page,
		"limit": limit,
	}
	response.SuccessResponse(c, http.StatusOK, "ok", data)
}

// ListIncomingRequests 入站好友请求列表
// @Summary 入站好友请求列表
// @Tags 好友
// @Security ApiKeyAuth
// @Produce json
// @Router /friends/requests/incoming [get]
func (h *FriendHandler) ListIncomingRequests(c *gin.Context) {
	payload, ok := c.Get(middleware.AuthorizationPayloadKey)
	if !ok {
		response.ErrorResponse(c, http.StatusUnauthorized, "未授权", nil)
		return
	}
	claims, ok := payload.(*service.JWTClaims)
	if !ok {
		response.ErrorResponse(c, http.StatusUnauthorized, "授权信息错误", nil)
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	status := c.Query("status")
	list, total, err := h.friendSvc.ListIncoming(c.Request.Context(), claims.UserID, model.FriendRequestStatus(status), page, limit)
	if err != nil {
		response.ErrorResponse(c, http.StatusInternalServerError, "获取入站请求失败", err.Error())
		return
	}
	data := map[string]any{
		"items": list,
		"total": total,
		"page":  page,
		"limit": limit,
	}
	response.SuccessResponse(c, http.StatusOK, "ok", data)
}

// ListOutgoingRequests 出站好友请求列表
// @Summary 出站好友请求列表
// @Tags 好友
// @Security ApiKeyAuth
// @Produce json
// @Router /friends/requests/outgoing [get]
func (h *FriendHandler) ListOutgoingRequests(c *gin.Context) {
	payload, ok := c.Get(middleware.AuthorizationPayloadKey)
	if !ok {
		response.ErrorResponse(c, http.StatusUnauthorized, "未授权", nil)
		return
	}
	claims, ok := payload.(*service.JWTClaims)
	if !ok {
		response.ErrorResponse(c, http.StatusUnauthorized, "授权信息错误", nil)
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	status := c.Query("status")
	list, total, err := h.friendSvc.ListOutgoing(c.Request.Context(), claims.UserID, model.FriendRequestStatus(status), page, limit)
	if err != nil {
		response.ErrorResponse(c, http.StatusInternalServerError, "获取出站请求失败", err.Error())
		return
	}
	data := map[string]any{
		"items": list,
		"total": total,
		"page":  page,
		"limit": limit,
	}
	response.SuccessResponse(c, http.StatusOK, "ok", data)
}

// UpdateRemark 更新好友备注
// @Summary 更新好友备注
// @Tags 好友
// @Security ApiKeyAuth
// @Accept json
// @Param friend_id path string true "好友ID"
// @Router /friends/remarks/{friend_id} [patch]
func (h *FriendHandler) UpdateRemark(c *gin.Context) {
	payload, ok := c.Get(middleware.AuthorizationPayloadKey)
	if !ok {
		response.ErrorResponse(c, http.StatusUnauthorized, "未授权", nil)
		return
	}
	claims, ok := payload.(*service.JWTClaims)
	if !ok {
		response.ErrorResponse(c, http.StatusUnauthorized, "授权信息错误", nil)
		return
	}
	fidStr := c.Param("friend_id")
	fid, ok := parseUUID(c, fidStr)
	if !ok { return }
	var body struct { Remark string `json:"remark" binding:"required"` }
	if err := c.ShouldBindJSON(&body); err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "请求参数错误", err.Error())
		return
	}
	if err := h.friendSvc.UpdateRemark(c.Request.Context(), claims.UserID, fid, body.Remark); err != nil {
		msg := err.Error()
		switch msg {
		case "备注长度不能超过50字符", "对方不是你的好友":
			response.ErrorResponse(c, http.StatusBadRequest, msg, nil)
		default:
			response.ErrorResponse(c, http.StatusInternalServerError, "更新备注失败", msg)
		}
		return
	}
	response.SuccessResponse(c, http.StatusOK, "备注已更新", gin.H{"friend_id": fidStr})
}

// DeleteFriend 删除好友
// @Summary 删除好友
// @Tags 好友
// @Security ApiKeyAuth
// @Param friend_id path string true "好友ID"
// @Router /friends/{friend_id} [delete]
func (h *FriendHandler) DeleteFriend(c *gin.Context) {
	payload, ok := c.Get(middleware.AuthorizationPayloadKey)
	if !ok {
		response.ErrorResponse(c, http.StatusUnauthorized, "未授权", nil)
		return
	}
	claims, ok := payload.(*service.JWTClaims)
	if !ok {
		response.ErrorResponse(c, http.StatusUnauthorized, "授权信息错误", nil)
		return
	}
	fidStr := c.Param("friend_id")
	fid, ok := parseUUID(c, fidStr)
	if !ok { return }
	if err := h.friendSvc.DeleteFriend(c.Request.Context(), claims.UserID, fid); err != nil {
		msg := err.Error()
		switch msg {
		case "不能删除自己":
			response.ErrorResponse(c, http.StatusBadRequest, msg, nil)
		default:
			response.ErrorResponse(c, http.StatusInternalServerError, "删除好友失败", msg)
		}
		return
	}
	response.SuccessResponse(c, http.StatusOK, "好友已删除", gin.H{"friend_id": fidStr})
}

// Block 拉黑
// @Summary 拉黑用户
// @Tags 好友
// @Security ApiKeyAuth
// @Param user_id path string true "用户ID"
// @Router /friends/blocks/{user_id} [post]
func (h *FriendHandler) Block(c *gin.Context) {
	payload, ok := c.Get(middleware.AuthorizationPayloadKey)
	if !ok {
		response.ErrorResponse(c, http.StatusUnauthorized, "未授权", nil)
		return
	}
	claims, ok := payload.(*service.JWTClaims)
	if !ok {
		response.ErrorResponse(c, http.StatusUnauthorized, "授权信息错误", nil)
		return
	}
	uidStr := c.Param("user_id")
	bid, ok := parseUUID(c, uidStr)
	if !ok { return }
	if err := h.friendSvc.Block(c.Request.Context(), claims.UserID, bid); err != nil {
		msg := err.Error()
		switch msg { case "不能拉黑自己":
			response.ErrorResponse(c, http.StatusBadRequest, msg, nil)
		default:
			response.ErrorResponse(c, http.StatusInternalServerError, "拉黑失败", msg)
		}
		return
	}
	response.SuccessResponse(c, http.StatusOK, "已拉黑该用户", gin.H{"user_id": uidStr})
}

// Unblock 取消拉黑
// @Summary 取消拉黑
// @Tags 好友
// @Security ApiKeyAuth
// @Param user_id path string true "用户ID"
// @Router /friends/blocks/{user_id} [delete]
func (h *FriendHandler) Unblock(c *gin.Context) {
	payload, ok := c.Get(middleware.AuthorizationPayloadKey)
	if !ok {
		response.ErrorResponse(c, http.StatusUnauthorized, "未授权", nil)
		return
	}
	claims, ok := payload.(*service.JWTClaims)
	if !ok {
		response.ErrorResponse(c, http.StatusUnauthorized, "授权信息错误", nil)
		return
	}
	uidStr := c.Param("user_id")
	bid, ok := parseUUID(c, uidStr)
	if !ok { return }
	if err := h.friendSvc.Unblock(c.Request.Context(), claims.UserID, bid); err != nil {
		response.ErrorResponse(c, http.StatusInternalServerError, "取消拉黑失败", err.Error())
		return
	}
	response.SuccessResponse(c, http.StatusOK, "已取消拉黑", gin.H{"user_id": uidStr})
}

// ListBlocks 黑名单
// @Summary 黑名单列表
// @Tags 好友
// @Security ApiKeyAuth
// @Produce json
// @Router /friends/blocks [get]
func (h *FriendHandler) ListBlocks(c *gin.Context) {
	payload, ok := c.Get(middleware.AuthorizationPayloadKey)
	if !ok {
		response.ErrorResponse(c, http.StatusUnauthorized, "未授权", nil)
		return
	}
	claims, ok := payload.(*service.JWTClaims)
	if !ok {
		response.ErrorResponse(c, http.StatusUnauthorized, "授权信息错误", nil)
		return
	}
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	list, total, err := h.friendSvc.ListBlocks(c.Request.Context(), claims.UserID, page, limit)
	if err != nil {
		response.ErrorResponse(c, http.StatusInternalServerError, "获取黑名单失败", err.Error())
		return
	}
	data := map[string]any{
		"items": list,
		"total": total,
		"page":  page,
		"limit": limit,
	}
	response.SuccessResponse(c, http.StatusOK, "ok", data)
}

// helper 解析UUID
func parseUUID(c *gin.Context, s string) (uuid.UUID, bool) {
	id, err := uuid.Parse(s)
	if err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "无效的ID", err.Error())
		return uuid.Nil, false
	}
	return id, true
}
