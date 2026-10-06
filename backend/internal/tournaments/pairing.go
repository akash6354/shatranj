package tournaments

func PairPlayers(players []Player) []Pairing {
	pairings := make([]Pairing, 0, (len(players)+1)/2)
	for index := 0; index < len(players); index += 2 {
		white := players[index]
		pairing := Pairing{WhiteID: white.UserID, WhiteScore: 0, Result: "*"}
		if index+1 == len(players) {
			pairing.Bye = true
			pairing.Result = "1-0"
			pairing.WhiteScore = 2
		} else {
			pairing.BlackID = players[index+1].UserID
		}
		pairings = append(pairings, pairing)
	}
	return pairings
}
