package account

import (
 "crypto/rand"
 "crypto/sha256"
 "encoding/hex"
 "encoding/json"
 "errors"
 "fmt"
 "os"
 "path/filepath"
 "strings"
 "sync"
 "fatewalker/game/character"
)

type CharacterRecord struct { ID string; *character.Character }
type Account struct { Username string; Salt string; Password string; Characters []CharacterRecord }
type Store struct { mu sync.Mutex; dir string }

func NewStore(dir string) *Store { return &Store{dir:dir} }
func validUsername(u string) bool {
 if len(u)<3||len(u)>24{return false}
 for _,r:=range u { if !(r>='a'&&r<='z'||r>='A'&&r<='Z'||r>='0'&&r<='9'||r=='_'||r=='-'){return false} }
 return true
}
func hash(password,salt string) string { h:=sha256.Sum256([]byte(salt+":"+password)); return hex.EncodeToString(h[:]) }
func (s *Store) path(username string) string { return filepath.Join(s.dir,strings.ToLower(username)+".json") }
func (s *Store) Register(username,password string) error {
 username=strings.TrimSpace(username)
 if !validUsername(username)||len(password)<8{return errors.New("username must be 3-24 letters/numbers/_/- and password must be at least 8 characters")}
 s.mu.Lock(); defer s.mu.Unlock()
 if err:=os.MkdirAll(s.dir,0700);err!=nil{return err}
 if _,err:=os.Stat(s.path(username));err==nil{return errors.New("that account already exists")}
 b:=make([]byte,16);if _,err:=rand.Read(b);err!=nil{return err};salt:=hex.EncodeToString(b)
 a:=&Account{Username:username,Salt:salt,Password:hash(password,salt),Characters:[]CharacterRecord{}}
 return s.writeLocked(a)
}
func (s *Store) Login(username,password string)(*Account,error){
 s.mu.Lock();defer s.mu.Unlock()
 data,err:=os.ReadFile(s.path(strings.TrimSpace(username)));if err!=nil{return nil,errors.New("invalid username or password")}
 var a Account
 if json.Unmarshal(data,&a)!=nil||a.Username==""||hash(password,a.Salt)!=a.Password{return nil,errors.New("invalid username or password")}
 return &a,nil
}
func (s *Store) Save(a *Account) error {s.mu.Lock();defer s.mu.Unlock();return s.writeLocked(a)}
func (s *Store) writeLocked(a *Account) error {
 data,err:=json.MarshalIndent(a,"","  ");if err!=nil{return err};if err:=os.MkdirAll(s.dir,0700);err!=nil{return err}
 tmp,err:=os.CreateTemp(s.dir,".account-*.tmp");if err!=nil{return err};name:=tmp.Name();defer os.Remove(name)
 if err:=tmp.Chmod(0600);err!=nil{tmp.Close();return err};if _,err:=tmp.Write(data);err!=nil{tmp.Close();return err};if err:=tmp.Sync();err!=nil{tmp.Close();return err};if err:=tmp.Close();err!=nil{return err}
 return os.Rename(name,s.path(a.Username))
}
func (a *Account) NewCharacter(name string)(*CharacterRecord,error){
 name=strings.TrimSpace(name);if len(name)<2||len(name)>24{return nil,errors.New("character name must be 2-24 characters")}
 for _,c:=range a.Characters{if strings.EqualFold(c.Name,name){return nil,errors.New("you already have a character with that name")}}
 b:=make([]byte,8);if _,err:=rand.Read(b);err!=nil{return nil,err}
 return &CharacterRecord{ID:hex.EncodeToString(b),Character:character.New(name)},nil
}
func (a *Account) DeleteCharacter(id string) error {
 for i,c:=range a.Characters{if c.ID==id{a.Characters=append(a.Characters[:i],a.Characters[i+1:]...);return nil}}
 return fmt.Errorf("character not found")
}
