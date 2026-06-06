package dto

type VaultMemberDTO struct {
	UID      int64  `json:"uid"`
	Nickname string `json:"nickname"`
	Role     string `json:"role"` // "owner" | "member"
}
