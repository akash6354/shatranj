package tournaments

import "testing"

func TestPairPlayersIncludesByeAndScore(t *testing.T) {
	players := []Player{
		{UserID: "a", Rating: 1800},
		{UserID: "b", Rating: 1700},
		{UserID: "c", Rating: 1600},
	}

	pairings := PairPlayers(players)
	if len(pairings) != 2 {
		t.Fatalf("pairing count = %d, want 2", len(pairings))
	}
	if pairings[0].WhiteID != "a" || pairings[0].BlackID != "b" || pairings[0].Result != "*" {
		t.Fatalf("first pairing = %+v, want a vs b in progress", pairings[0])
	}
	bye := pairings[1]
	if !bye.Bye || bye.WhiteID != "c" || bye.BlackID != "" || bye.Result != "1-0" || bye.WhiteScore != 2 {
		t.Fatalf("bye = %+v, want c to receive a 2-point bye", bye)
	}
}

func TestSortStandingsAssignsStableRanks(t *testing.T) {
	players := []Player{
		{UserID: "b", Score: 2, Rating: 1500},
		{UserID: "c", Score: 2, Rating: 1600},
		{UserID: "a", Score: 2, Rating: 1600},
		{UserID: "d", Score: 1, Rating: 2200},
	}

	SortStandings(players)
	wantIDs := []string{"a", "c", "b", "d"}
	for index, want := range wantIDs {
		if players[index].UserID != want || players[index].Rank != index+1 {
			t.Fatalf("standing[%d] = %+v, want user %q rank %d", index, players[index], want, index+1)
		}
	}
}
