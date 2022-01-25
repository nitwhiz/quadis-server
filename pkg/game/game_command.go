package game

const CommandLeft = "L"
const CommandRight = "R"
const CommandDown = "D"
const CommandRotate = "X"
const CommandHardLock = "P"
const CommandHold = "H"

func (g *Game) Command(cmd string) {
	g.mu.Lock()
	defer g.mu.Unlock()

	g.FallingPiece.mu.Lock()
	defer g.FallingPiece.mu.Unlock()

	// todo: piece locking runs into deadlock (?)

	switch cmd {
	case CommandLeft:
		g.moveFallingPiece(0, -1, 0)
		return
	case CommandRight:
		g.moveFallingPiece(0, 1, 0)
		return
	case CommandDown:
		g.moveFallingPiece(0, 0, 1)
		return
	case CommandRotate:
		g.moveFallingPiece(1, 0, 0)
		return
	case CommandHardLock:
		g.hardLockFallingPiece()
		return
	case CommandHold:
		g.holdFallingPiece()
		return
	default:
		return
	}
}
