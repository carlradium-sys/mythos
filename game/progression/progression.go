package progression
func XPForLevel(l int) int { if l<=1{return 0}; return 100*(l-1)*(l-1) }
func LevelFromXP(x int) int { l:=1; for l<50&&x>=XPForLevel(l+1){l++}; return l }
func XPToNextLevel(l,x int) int { n:=XPForLevel(l+1)-x; if n<0{return 0}; return n }
