package model

import "math/rand"

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
	if len(W) == 0 {
		W = append(W, WaitingList{
			ID: id,
		})
	} else {
		for i := 0; i < len(W); i++ {
			if W[i].ID != id && i != len(W)-1 {
				break
			} else if W[i].ID != id && i == len(W)-1 {
				W = append(W, WaitingList{
					ID: id,
				})
			} else {
				break
			}
		}
	}

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
	if len(W) != 0 && len(W) > 2 {
		one := rand.Intn(len(W))
		oneID = W[one].ID
		W = append(W[:one], W[one+1:]...)

		two := rand.Intn(len(W))
		twoID = W[two].ID
		W = append(W[:two], W[two+1:]...)

		return oneID, twoID
	} else if len(W) == 2 {
		oneID = W[0].ID
		W = append(W[:0], W[1:]...)

		twoID = W[0].ID
		W = append(W[:0], W[1:]...)

		return oneID, twoID
	} else {
		return 0, 0
	}
}

func AddToRoom(oneID, twoID int64) {
	for i := 0; i < len(U); i++ {
		if U[i].ID == oneID || U[i].ID == twoID {
			U[i].Index = "chatting"
		}
	}

	R = append(R, Room{
		UOne: oneID,
		UTwo: twoID,
	})
}

func RestartRoom(oneID int64) int64 {
	var twoID int64
	for i := 0; i < len(R); i++ {
		if R[i].UOne == oneID || R[i].UTwo == oneID {
			if R[i].UOne != oneID {
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

func RoomWriter() {
	for {
		if len(R) == 1 {
			if R[0].UOne == R[0].UTwo {
				R = []Room{}
			}
		} else if len(R) != 0 {
			for i := 0; i < len(R); i++ {
				if R[i].UOne == R[i].UTwo {
					R = append(R[:i], R[i+1:]...)
				}
			}
		}
	}
}

func RoomUser(oneID int64) int64 {
	for i := 0; i < len(R); i++ {
		if R[i].UOne == oneID {
			return R[i].UTwo
		} else if R[i].UTwo == oneID {
			return R[i].UOne
		}
	}

	return 0
}

func DeleteRoom(oneID int64) int64 {
	var twoID int64
	for i := 0; i < len(R); i++ {
		if R[i].UOne == oneID || R[i].UTwo == oneID {
			if R[i].UOne != oneID {
				twoID = R[i].UOne
				AddToWaitingList(R[i].UOne)
			} else {
				twoID = R[i].UTwo
				AddToWaitingList(R[i].UTwo)
			}
			R = append(R[:i], R[i+1:]...)
			i = len(R)
		}
	}

	return twoID
}
