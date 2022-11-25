package model

type WaitingList struct {
	ID int64 `json:"id"`
}

type Room struct {
	UOne int64 `json:"u_one"`
	UTwo int64 `json:"u_two"`
}

type User struct {
	ID    int64  `json:"id"`
	Index string `json:"index"`
}

type Update struct {
	UpdateId int     `json:"update_id"`
	Message  Message `json:"message"`
}

type Message struct {
	Chat Chat   `json:"chat"`
	Text string `json:"text"`
}

type Chat struct {
	ChatId int64 `json:"id"`
}

type RestResponse struct {
	Result []Update `json:"result"`
}

type BotMessage struct {
	ChatId int64  `json:"chat_id"`
	Text   string `json:"text"`
}

var U []User
var W []WaitingList
var R []Room

func GetUser(id int64) User {
	for i := 0; i < len(U); i++ {
		if U[i].ID == id {
			return U[i]
		}
	}

	U = append(U, User{
		ID:    id,
		Index: "home",
	})

	return User{
		ID:    id,
		Index: "home",
	}
}

func UpdateUser(id int64, index string) {
	for i := 0; i < len(U); i++ {
		if U[i].ID == id {
			U[i].Index = index
			i = len(U)
		}
	}
}

func AddToWaitingList(id int64) {
	W = append(W, WaitingList{
		ID: id,
	})
}

func DeleteFromWaitingList(id int64) {
	for i := 0; i < len(W); i++ {
		if W[i].ID == id {
			W = append(W[:i], W[i+1:]...)
		}
	}
}

func GetFromWaitingList() (int64, int64) {
	var oneID int64
	var twoID int64
	if len(W) != 0 && len(W) >= 2 {
		oneID = W[0].ID
		W = append(W[:0], W[1:]...)

		twoID = W[0].ID
		W = append(W[:0], W[1:]...)
	} else {
		return 0, 0
	}

	return oneID, twoID
}

func AddToRoom(oneID, twoID int64) {
	for i := 0; i < len(U); i++ {
		if U[i].ID == oneID || U[i].ID == twoID {
			U[i].Index = "chatting"
			i = len(U)
		}
	}

	R = append(R, Room{
		UOne: oneID,
		UTwo: twoID,
	})
}

func DeleteFromRoom(id int64) int64 {
	var twoID int64
	for i := 0; i < len(R); i++ {
		if R[i].UOne == id || R[i].UTwo == id {
			if R[i].UOne != id {
				twoID = R[i].UOne
			} else {
				twoID = R[i].UTwo
			}
			AddToWaitingList(R[i].UOne)
			AddToWaitingList(R[i].UTwo)
			R = append(R[:i], R[i+1:]...)
			i = len(R)
		}
	}

	return twoID
}
