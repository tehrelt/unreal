package dto

import "github.com/tehrelt/unreal/internal/entity"

type FetchMessagesDto struct {
	Mailbox string
	Limit   int
	Page    int
}

type FetchedMessagesDto struct {
	Messages    []entity.Message
	HasNextPage bool
	Total       int
}
