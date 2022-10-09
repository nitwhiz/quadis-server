package game

type Command string

const CommandLeft = Command("L")
const CommandRight = Command("R")
const CommandDown = Command("D")
const CommandRotate = Command("X")
const CommandHardLock = Command("P")
const CommandHold = Command("H")

func (g *Game) HandleCommand(cmd Command) {
	switch cmd {
	case CommandLeft:
		g.tryTranslateFallingPiece(0, -1, 0)
		break
	case CommandRight:
		g.tryTranslateFallingPiece(0, 1, 0)
		break
	case CommandDown:
		g.tryTranslateFallingPiece(0, 0, 1)
		break
	case CommandRotate:
		g.tryTranslateFallingPiece(1, 0, 0)
		break
	case CommandHardLock:
		g.hardLockFallingPiece()
		break
	case CommandHold:
		g.tryHoldFallingPiece()
		break
	default:
		break
	}
}
