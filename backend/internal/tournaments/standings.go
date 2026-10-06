package tournaments

import "sort"

func SortStandings(players []Player) {
	sort.SliceStable(players, func(i, j int) bool {
		if players[i].Score != players[j].Score {
			return players[i].Score > players[j].Score
		}
		if players[i].Rating != players[j].Rating {
			return players[i].Rating > players[j].Rating
		}
		return players[i].UserID < players[j].UserID
	})
	for index := range players {
		players[index].Rank = index + 1
	}
}
