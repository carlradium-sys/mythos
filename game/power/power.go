package power
type Domain string
const(Storm Domain="Storm";Tide Domain="Tide";Ember Domain="Ember";Aegis Domain="Aegis")
type Power struct{Name string;Domain Domain;UnlockLevel,ManaCost,Damage int;Description string}
func ForDomain(d Domain)[]Power{switch d{
case Storm:return[]Power{{"Stormspark",d,5,8,30,"A crackling bolt leaps from your hand."},{"Thunderfall",d,8,16,55,"Thunder answers your call and strikes from above."}}
case Tide:return[]Power{{"Undertow",d,5,8,28,"A crushing ribbon of water coils around your foe."},{"Leviathan Surge",d,8,16,52,"A towering wave crashes through the battlefield."}}
case Ember:return[]Power{{"Cinderbrand",d,5,8,32,"White-hot flame brands the target."},{"Promethean Star",d,8,16,58,"A miniature sun blooms at your fingertips."}}
case Aegis:return[]Power{{"Aegis Pulse",d,5,8,20,"Divine force erupts outward from your shield."},{"Olympian Bulwark",d,8,16,48,"A radiant wall crashes into your enemy."}}};return nil}
