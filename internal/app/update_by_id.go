package app

func updateByID[T any](items []T, id int, itemID func(*T) int, update func(*T)) bool {
	for i := range items {
		if itemID(&items[i]) == id {
			update(&items[i])
			return true
		}
	}
	return false
}

func updateByIDAndMarkDirty[T any](g *Game, items []T, id int, itemID func(*T) int, update func(*T)) {
	if updateByID(items, id, itemID, update) {
		g.markDirty()
	}
}
