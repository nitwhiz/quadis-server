package event

const originRoom = "room"
const originGame = "game"
const originSystem = "system"

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

func OriginSystem() *Origin {
	return &Origin{
		Id:   "",
		Type: originSystem,
	}
}
