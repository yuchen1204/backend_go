package service

import (
	"backend/internal/model"
	"backend/internal/repository"
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
)

// FriendService 好友相关服务

type FriendService interface {
	CreateRequest(ctx context.Context, requesterID, receiverID uuid.UUID, note string) (*model.FriendRequest, error)
	ListIncoming(ctx context.Context, userID uuid.UUID, status model.FriendRequestStatus, page, limit int) ([]model.FriendRequest, int64, error)
	ListOutgoing(ctx context.Context, userID uuid.UUID, status model.FriendRequestStatus, page, limit int) ([]model.FriendRequest, int64, error)
	AcceptRequest(ctx context.Context, requestID, actorID uuid.UUID) error
	RejectRequest(ctx context.Context, requestID, actorID uuid.UUID) error
	CancelRequest(ctx context.Context, requestID, actorID uuid.UUID) error
	ListFriends(ctx context.Context, userID uuid.UUID, search string, page, limit int) ([]model.Friendship, int64, error)
	DeleteFriend(ctx context.Context, userID, friendID uuid.UUID) error
	UpdateRemark(ctx context.Context, userID, friendID uuid.UUID, remark string) error
	Block(ctx context.Context, userID, blockedID uuid.UUID) error
	Unblock(ctx context.Context, userID, blockedID uuid.UUID) error
	ListBlocks(ctx context.Context, userID uuid.UUID, page, limit int) ([]model.BlockList, int64, error)
}

type friendService struct {
	friendReqRepo repository.FriendRequestRepository
	friendRepo    repository.FriendshipRepository
	blockRepo     repository.BlockListRepository
	banRepo       repository.FriendBanRepository
	userRepo      repository.UserRepository
	rateLimitRepo repository.RateLimitRepository
	mailSvc       MailService
	userLogSvc    UserActionLogService
	maxDailyReq   int
	maxFriends    int
	chatRoomRepo  repository.ChatRoomRepository
}

func NewFriendService(friendReqRepo repository.FriendRequestRepository,
	friendRepo repository.FriendshipRepository,
	blockRepo repository.BlockListRepository,
	banRepo repository.FriendBanRepository,
	userRepo repository.UserRepository,
	rateLimitRepo repository.RateLimitRepository,
	mailSvc MailService,
	userLogSvc UserActionLogService,
	maxDailyReq int,
	maxFriends int,
	chatRoomRepo repository.ChatRoomRepository,
) FriendService {
	return &friendService{
		friendReqRepo: friendReqRepo,
		friendRepo:    friendRepo,
		blockRepo:     blockRepo,
		banRepo:       banRepo,
		userRepo:      userRepo,
		rateLimitRepo: rateLimitRepo,
		mailSvc:       mailSvc,
		userLogSvc:    userLogSvc,
		maxDailyReq:   maxDailyReq,
		maxFriends:    maxFriends,
		chatRoomRepo:  chatRoomRepo,
	}
}

func (s *friendService) CreateRequest(ctx context.Context, requesterID, receiverID uuid.UUID, note string) (*model.FriendRequest, error) {
	if requesterID == receiverID {
		return nil, errors.New("不能向自己发起请求")
	}
	if len(note) > 200 {
		return nil, errors.New("申请备注长度不能超过200字符")
	}

	// 禁用检查（管理员禁用发起申请）
	if _, err := s.banRepo.GetActiveBan(requesterID, time.Now()); err == nil {
		return nil, errors.New("您当前被禁用发起好友申请")
	}

	// 拉黑检查（任一方向被拉黑都不允许）
	if blocked, err := s.blockRepo.IsBlocked(receiverID, requesterID); err != nil {
		return nil, err
	} else if blocked {
		return nil, errors.New("对方已将你拉黑，无法发起请求")
	}
	if blocked, err := s.blockRepo.IsBlocked(requesterID, receiverID); err != nil {
		return nil, err
	} else if blocked {
		return nil, errors.New("请先取消对该用户的拉黑")
	}

	// 每日请求频率限制（按用户计数，复用 RateLimitRepository）
	key := "friendreq:" + requesterID.String()
	if count, err := s.rateLimitRepo.Increment(ctx, key); err != nil {
		return nil, err
	} else if int(count) > s.maxDailyReq {
		return nil, errors.New("请求过于频繁，请明天再试")
	}

	// 好友上限检查（双方）
	if c, err := s.friendRepo.CountByUser(requesterID); err != nil {
		return nil, err
	} else if int(c) >= s.maxFriends {
		return nil, errors.New("您的好友数已达上限")
	}
	if c, err := s.friendRepo.CountByUser(receiverID); err != nil {
		return nil, err
	} else if int(c) >= s.maxFriends {
		return nil, errors.New("对方的好友数已达上限")
	}

	// 已是好友检查
	if exists, err := s.friendRepo.Exists(requesterID, receiverID); err != nil {
		return nil, err
	} else if exists {
		return nil, errors.New("你们已经是好友")
	}

	// 待处理请求幂等：同向、反向
	if _, err := s.friendReqRepo.FindPending(requesterID, receiverID); err == nil {
		return nil, errors.New("已存在待处理的请求")
	}
	if _, err := s.friendReqRepo.FindPending(receiverID, requesterID); err == nil {
		return nil, errors.New("对方已有向你的待处理请求")
	}

	fr := &model.FriendRequest{
		RequesterID: requesterID,
		ReceiverID:  receiverID,
		Status:      model.FriendRequestPending,
		Note:        note,
	}
	if err := s.friendReqRepo.Create(fr); err != nil {
		return nil, err
	}

	// 行为日志（发起请求）
	_ = s.userLogSvc.Create(ctx, &model.UserActionLog{
		UserID:   &requesterID,
		Action:   "friend_request_create",
		Details:  "to:" + receiverID.String(),
	})

	// 发送邮件通知给接收方（异步）
	go func() {
		// 查询接收方邮箱与请求者展示名
		receiver, err1 := s.userRepo.GetByID(receiverID)
		requester, err2 := s.userRepo.GetByID(requesterID)
		if err1 != nil || err2 != nil || receiver.Email == "" {
			return
		}
		requesterName := requester.Nickname
		if requesterName == "" {
			requesterName = requester.Username
		}
		receiverName := receiver.Nickname
		if receiverName == "" {
			receiverName = receiver.Username
		}
		_ = s.mailSvc.SendFriendRequestNotification(
			receiver.Email,
			requesterName,
			requester.Username,
			receiverName,
			note,
			fr.CreatedAt,
		)
	}()

	return fr, nil
}

func (s *friendService) ListIncoming(ctx context.Context, userID uuid.UUID, status model.FriendRequestStatus, page, limit int) ([]model.FriendRequest, int64, error) {
	return s.friendReqRepo.ListIncoming(userID, status, page, limit)
}

func (s *friendService) ListOutgoing(ctx context.Context, userID uuid.UUID, status model.FriendRequestStatus, page, limit int) ([]model.FriendRequest, int64, error) {
	return s.friendReqRepo.ListOutgoing(userID, status, page, limit)
}

func (s *friendService) AcceptRequest(ctx context.Context, requestID, actorID uuid.UUID) error {
    // 获取请求
    fr, err := s.friendReqRepo.GetByID(requestID)
    if err != nil {
        return errors.New("请求不存在")
    }
    if fr.Status != model.FriendRequestPending {
        return errors.New("请求已被处理")
    }
    // 权限：只有接收方可以接受
    if fr.ReceiverID != actorID {
        return errors.New("无权处理该请求")
    }

    // 拉黑/禁用/上限/已是好友检查
    if _, err := s.banRepo.GetActiveBan(actorID, time.Now()); err == nil {
        return errors.New("您当前被禁用发起好友申请")
    }
    if blocked, err := s.blockRepo.IsBlocked(fr.RequesterID, fr.ReceiverID); err != nil {
        return err
    } else if blocked {
        return errors.New("双方存在拉黑关系，无法成为好友")
    }
    if blocked, err := s.blockRepo.IsBlocked(fr.ReceiverID, fr.RequesterID); err != nil {
        return err
    } else if blocked {
        return errors.New("双方存在拉黑关系，无法成为好友")
    }
    if exists, err := s.friendRepo.Exists(fr.RequesterID, fr.ReceiverID); err != nil {
        return err
    } else if exists {
        // 幂等：若已是好友，仍补齐双向记录，防止历史只存在单向记录导致列表为空
        if err := s.friendRepo.CreateIgnoreDuplicate(&model.Friendship{UserID: fr.RequesterID, FriendID: fr.ReceiverID}); err != nil {
            return err
        }
        if err := s.friendRepo.CreateIgnoreDuplicate(&model.Friendship{UserID: fr.ReceiverID, FriendID: fr.RequesterID}); err != nil {
            return err
        }
        // 将请求标记为 accepted
        now := time.Now()
        fr.Status = model.FriendRequestAccepted
        fr.HandledAt = &now
        if err := s.friendReqRepo.Update(fr); err != nil {
            return err
        }
        return nil
    }
    if c, err := s.friendRepo.CountByUser(fr.RequesterID); err != nil {
        return err
    } else if int(c) >= s.maxFriends {
        return errors.New("对方的好友数已达上限")
    }
    if c, err := s.friendRepo.CountByUser(fr.ReceiverID); err != nil {
        return err
    } else if int(c) >= s.maxFriends {
        return errors.New("您的好友数已达上限")
    }

    // TODO: 使用数据库事务保证原子性
    // 创建双向好友（忽略重复，避免并发/历史数据导致的唯一约束冲突）
    if err := s.friendRepo.CreateIgnoreDuplicate(&model.Friendship{UserID: fr.RequesterID, FriendID: fr.ReceiverID}); err != nil {
        return err
    }
    if err := s.friendRepo.CreateIgnoreDuplicate(&model.Friendship{UserID: fr.ReceiverID, FriendID: fr.RequesterID}); err != nil {
        return err
    }

    // 更新请求状态
    now := time.Now()
    fr.Status = model.FriendRequestAccepted
    fr.HandledAt = &now
    if err := s.friendReqRepo.Update(fr); err != nil {
        return err
    }

    // 行为日志
    _ = s.userLogSvc.Create(ctx, &model.UserActionLog{
        UserID:  &actorID,
        Action:  "friend_request_accept",
        Details: "request:" + requestID.String(),
    })

    // 邮件通知：告知发起方（requester）结果，包含请求和处理时间
    go func() {
        requester, err1 := s.userRepo.GetByID(fr.RequesterID)
        receiver, err2 := s.userRepo.GetByID(fr.ReceiverID)
        if err1 != nil || err2 != nil || requester.Email == "" {
            return
        }
        otherName := receiver.Nickname
        if otherName == "" { otherName = receiver.Username }
        _ = s.mailSvc.SendFriendRequestResultNotification(
            requester.Email,
            otherName,
            receiver.Username,
            "accepted",
            fr.CreatedAt,
            *fr.HandledAt,
        )
    }()
    return nil
}

func (s *friendService) RejectRequest(ctx context.Context, requestID, actorID uuid.UUID) error {
    fr, err := s.friendReqRepo.GetByID(requestID)
    if err != nil {
        return errors.New("请求不存在")
    }
    if fr.Status != model.FriendRequestPending {
        return errors.New("请求已被处理")
    }
    if fr.ReceiverID != actorID {
        return errors.New("无权处理该请求")
    }
    now := time.Now()
    fr.Status = model.FriendRequestRejected
    fr.HandledAt = &now
    if err := s.friendReqRepo.Update(fr); err != nil {
        return err
    }
    _ = s.userLogSvc.Create(ctx, &model.UserActionLog{
        UserID:  &actorID,
        Action:  "friend_request_reject",
        Details: "request:" + requestID.String(),
    })
    // 邮件通知：告知发起方（requester）被拒绝
    go func() {
        requester, err1 := s.userRepo.GetByID(fr.RequesterID)
        receiver, err2 := s.userRepo.GetByID(fr.ReceiverID)
        if err1 != nil || err2 != nil || requester.Email == "" || fr.HandledAt == nil {
            return
        }
        otherName := receiver.Nickname
        if otherName == "" { otherName = receiver.Username }
        _ = s.mailSvc.SendFriendRequestResultNotification(
            requester.Email,
            otherName,
            receiver.Username,
            "rejected",
            fr.CreatedAt,
            *fr.HandledAt,
        )
    }()
    return nil
}

func (s *friendService) CancelRequest(ctx context.Context, requestID, actorID uuid.UUID) error {
    fr, err := s.friendReqRepo.GetByID(requestID)
    if err != nil {
        return errors.New("请求不存在")
    }
    if fr.Status != model.FriendRequestPending {
        return errors.New("请求已被处理")
    }
    if fr.RequesterID != actorID {
        return errors.New("无权撤回该请求")
    }
    now := time.Now()
    fr.Status = model.FriendRequestCanceled
    fr.HandledAt = &now
    if err := s.friendReqRepo.Update(fr); err != nil {
        return err
    }
    _ = s.userLogSvc.Create(ctx, &model.UserActionLog{
        UserID:  &actorID,
        Action:  "friend_request_cancel",
        Details: "request:" + requestID.String(),
    })
    // 邮件通知：告知接收方（receiver）对方已撤回
    go func() {
        requester, err1 := s.userRepo.GetByID(fr.RequesterID)
        receiver, err2 := s.userRepo.GetByID(fr.ReceiverID)
        if err1 != nil || err2 != nil || receiver.Email == "" || fr.HandledAt == nil {
            return
        }
        otherName := requester.Nickname
        if otherName == "" { otherName = requester.Username }
        _ = s.mailSvc.SendFriendRequestResultNotification(
            receiver.Email,
            otherName,
            requester.Username,
            "cancelled",
            fr.CreatedAt,
            *fr.HandledAt,
        )
    }()
    return nil
}

func (s *friendService) ListFriends(ctx context.Context, userID uuid.UUID, search string, page, limit int) ([]model.Friendship, int64, error) {
	return s.friendRepo.List(userID, search, page, limit)
}

func (s *friendService) DeleteFriend(ctx context.Context, userID, friendID uuid.UUID) error {
    if userID == friendID {
        return errors.New("不能删除自己")
    }
    // 仅当存在好友关系时删除（幂等处理）
    exists, err := s.friendRepo.Exists(userID, friendID)
    if err != nil {
        return err
    }
    if !exists {
        // 幂等：无操作
        return nil
    }
    if err := s.friendRepo.DeletePair(userID, friendID); err != nil {
        return err
    }
    // 尝试将双方的聊天房间设为 inactive（忽略错误以保证删除流程）
    if s.chatRoomRepo != nil {
        _ = s.chatRoomRepo.DeactivateByUsers(userID, friendID)
    }
    _ = s.userLogSvc.Create(ctx, &model.UserActionLog{
        UserID:  &userID,
        Action:  "friend_delete",
        Details: "friend:" + friendID.String(),
    })
    return nil
}

func (s *friendService) UpdateRemark(ctx context.Context, userID, friendID uuid.UUID, remark string) error {
    if len(remark) > 50 {
        return errors.New("备注长度不能超过50字符")
    }
    // 必须是现有好友
    exists, err := s.friendRepo.Exists(userID, friendID)
    if err != nil {
        return err
    }
    if !exists {
        return errors.New("对方不是你的好友")
    }
    if err := s.friendRepo.UpdateRemark(userID, friendID, remark); err != nil {
        return err
    }
    _ = s.userLogSvc.Create(ctx, &model.UserActionLog{
        UserID:  &userID,
        Action:  "friend_remark_update",
        Details: "friend:" + friendID.String(),
    })
    return nil
}

func (s *friendService) Block(ctx context.Context, userID, blockedID uuid.UUID) error {
	if userID == blockedID {
		return errors.New("不能拉黑自己")
	}
	return s.blockRepo.Block(userID, blockedID)
}

func (s *friendService) Unblock(ctx context.Context, userID, blockedID uuid.UUID) error {
	return s.blockRepo.Unblock(userID, blockedID)
}

func (s *friendService) ListBlocks(ctx context.Context, userID uuid.UUID, page, limit int) ([]model.BlockList, int64, error) {
    return s.blockRepo.List(userID, page, limit)
}
