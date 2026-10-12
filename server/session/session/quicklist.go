package session

import (
	"strconv"
	"strings"

	"fatewalker/game/power"
)

func (s *Session) quicklist(args []string) {
	if s.Character == nil { return }
	if len(args) == 0 || strings.EqualFold(args[0], "list") {
		s.showQuicklist()
		return
	}
	switch strings.ToLower(args[0]) {
	case "add":
		if len(args) < 2 { s.WriteLine("Add which power? Use: ql add <power>"); return }
		if s.Character.Domain == "" { s.WriteLine("Awaken a divine domain before adding powers to your quicklist."); return }
		name := strings.Join(args[1:], " ")
		for _, p := range power.Unlocked(s.Character.Domain, s.Character.Level) {
			if strings.EqualFold(p.Name, name) || strings.EqualFold(strings.ReplaceAll(p.Name, " ", ""), strings.ReplaceAll(name, " ", "")) {
				for _, saved := range s.Character.Quicklist {
					if strings.EqualFold(saved, p.Name) { s.WriteLine("%s is already on your quicklist.", p.Name); return }
				}
				s.Character.Quicklist = append(s.Character.Quicklist, p.Name)
				s.WriteLine("%s added to quicklist slot %d.", p.Name, len(s.Character.Quicklist))
				return
			}
		}
		s.WriteLine("That power is not currently unlocked for your domain. Use 'powers' to check your available powers.")
	case "remove", "delete", "del":
		if len(args) != 2 { s.WriteLine("Remove which slot? Use: ql remove <slot>"); return }
		n, err := strconv.Atoi(args[1])
		if err != nil || n < 1 || n > len(s.Character.Quicklist) { s.WriteLine("That quicklist slot does not exist."); return }
		removed := s.Character.Quicklist[n-1]
		s.Character.Quicklist = append(s.Character.Quicklist[:n-1], s.Character.Quicklist[n:]...)
		s.WriteLine("Removed %s from quicklist.", removed)
	case "clear":
		s.Character.Quicklist = nil
		s.WriteLine("Combat quicklist cleared.")
	default:
		n, err := strconv.Atoi(args[0])
		if err != nil { s.WriteLine("Use 'ql' to view your list, 'ql add <power>' to add a power, or 'ql <slot> <target>' to cast."); return }
		if n < 1 || n > len(s.Character.Quicklist) { s.WriteLine("That quicklist slot does not exist. Use 'ql' to view your list."); return }
		if len(args) < 2 { s.WriteLine("Choose a target: ql %d <target>", n); return }
		if s.Enemy == nil || s.Enemy.HP <= 0 { s.WriteLine("There is no enemy here to target."); return }
		target := strings.Join(args[1:], " ")
		if !strings.EqualFold(target, s.Enemy.Name) && !strings.EqualFold(target, "target") && !strings.EqualFold(target, "enemy") {
			s.WriteLine("You cannot target %s here. Your current enemy is %s.", target, s.Enemy.Name)
			return
		}
		s.cast([]string{s.Character.Quicklist[n-1]})
	}
}

func (s *Session) showQuicklist() {
	s.WriteLine("\x1b[1;33mCOMBAT QUICKLIST\x1b[0m")
	if len(s.Character.Quicklist) == 0 {
		s.WriteLine("Your quicklist is empty. Add an unlocked power with: ql add <power>")
		return
	}
	for i, name := range s.Character.Quicklist { s.WriteLine("  %d. %s", i+1, name) }
	s.WriteLine("Cast with: ql <slot> <target>  |  Add: ql add <power>  |  Remove: ql remove <slot>")
}
