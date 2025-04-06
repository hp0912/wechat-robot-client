package vars

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"math/rand"
	"time"
)

type Robot struct{}

func (r *Robot) CreateDeviceName() string {
	firstNames := []string{
		"Oliver", "Emma", "Liam", "Ava", "Noah", "Sophia", "Elijah", "Isabella",
		"James", "Mia", "William", "Amelia", "Benjamin", "Harper", "Lucas", "Evelyn",
		"Henry", "Abigail", "Alexander", "Ella", "Jackson", "Scarlett", "Sebastian",
		"Grace", "Aiden", "Chloe", "Matthew", "Zoey", "Samuel", "Lily", "David",
		"Aria", "Joseph", "Riley", "Carter", "Nora", "Owen", "Luna", "Daniel",
		"Sofia", "Gabriel", "Ellie", "Matthew", "Avery", "Isaac", "Mila", "Leo",
		"Julian", "Layla",
	}

	lastNames := []string{
		"Smith", "Johnson", "Williams", "Brown", "Jones", "Garcia", "Miller", "Davis",
		"Rodriguez", "Martinez", "Hernandez", "Lopez", "Gonzalez", "Wilson", "Anderson",
		"Thomas", "Taylor", "Moore", "Jackson", "Martin", "Lee", "Perez", "Thompson",
		"White", "Harris", "Sanchez", "Clark", "Ramirez", "Lewis", "Robinson", "Walker",
		"Young", "Allen", "King", "Wright", "Scott", "Torres", "Nguyen", "Hill",
		"Flores", "Green", "Adams", "Nelson", "Baker", "Hall", "Rivera", "Campbell",
		"Mitchell", "Carter", "Roberts", "Gomez", "Phillips", "Evans",
	}

	source := rand.NewSource(time.Now().UnixNano())
	rng := rand.New(source)
	firstName := firstNames[rng.Intn(len(firstNames))]
	lastName := lastNames[rng.Intn(len(lastNames))]

	return fmt.Sprintf("%s %s's Pad", firstName, lastName)
}

func (r *Robot) CreateDeviceID(s string) string {
	if s == "" || s == "string" {
		var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
		b := make([]rune, 15)
		for i := range b {
			b[i] = letters[rand.Intn(len(letters))]
		}
		s = string(b)
	}

	hash := md5.Sum([]byte(s))
	return "49" + hex.EncodeToString(hash[:])[2:]
}
