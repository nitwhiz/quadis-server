package event

const originRoom = "room"
const originGame = "game"

func OriginRoom(id string) *Origin {
	return &Origin{
		Id:   id,
		Type: originRoom,
	}
}

func OriginGame(id string) *Origin {
	return &Origin{
		Id:   id,
		Type: originGame,
	}
}
