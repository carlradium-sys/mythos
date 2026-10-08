package item
type Tier int
const(Common Tier=iota;Uncommon;Rare;Epic;Legendary;Mythic)
type Item struct{Name string;Tier Tier;Damage int;Kind string}
func(i Item)TierName()string{n:=[]string{"Common","Uncommon","Rare","Epic","Legendary","Mythic"};if int(i.Tier)<len(n){return n[i.Tier]};return"Unknown"}
