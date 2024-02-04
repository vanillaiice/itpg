package itpg

import (
	"testing"
	"time"
)

func initAuthDB(path ...string) (*AuthDB, error) {
	if len(path) == 0 {
		path = append(path, ":memory:")
	}
	authDB, err := NewAuthDB(path[0])
	if err != nil {
		return nil, err
	}

	err = authDB.AddUser("giorno-giovanna", "goldenwind")
	if err != nil {
		return nil, err
	}
	err = authDB.AddUser("jp-polnareff", "silverchariot")
	if err != nil {
		return nil, err
	}
	err = authDB.AddUser("dio-brando", "theworld")
	if err != nil {
		return nil, err
	}

	err = authDB.AddSession("giorno-giovanna", "goldenexperience", time.Now())
	if err != nil {
		return nil, err
	}
	err = authDB.AddSession("jp-polnareff", "horahorahora", time.Now().Add(time.Minute))
	if err != nil {
		return nil, err
	}

	return authDB, err
}

func TestNewAuthDB(t *testing.T) {
	db, err := NewAuthDB(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	db.Close()
}

func TestAddUser(t *testing.T) {
	db, err := initAuthDB()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	err = db.AddUser("fugo-pannacotta", "purplehaze")
	if err != nil {
		t.Error(err)
	}
	err = db.AddUser("fugo-pannacotta", "purplehaze")
	if err == nil {
		t.Error("expected failure")
	}
}

func TestAddSession(t *testing.T) {
	db, err := initAuthDB()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	err = db.AddSession("dio-brando", "mudamudamuda", time.Now().Add(10*time.Second))
	if err != nil {
		t.Error(err)
	}
	err = db.AddSession("fugo-pannacotta", "burbleha$e", time.Now().Add(10*time.Second))
	if err == nil {
		t.Error("expected failure")
	}
}

func TestUserExists(t *testing.T) {
	db, err := initAuthDB()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	exists, err := db.UserExists("jotaro-kujo")
	if err != nil {
		t.Error(err)
	}
	if exists {
		t.Errorf("got %v, want %v", exists, false)
	}
	exists, err = db.UserExists("giorno-giovanna")
	if err != nil {
		t.Error(err)
	}
	if !exists {
		t.Errorf("got %v, want %v", exists, false)
	}
}

func TestSessionExists(t *testing.T) {
	db, err := initAuthDB()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	exists, err := db.SessionExists("goldenexperience")
	if err != nil {
		t.Error(err)
	}
	if !exists {
		t.Errorf("got %v, want %v", exists, true)
	}
	exists, err = db.SessionExists("silverexperience")
	if err != nil {
		t.Error(err)
	}
	if exists {
		t.Errorf("got %v, want %v", exists, false)
	}
}

func TestSessionExistsByUsername(t *testing.T) {
	db, err := initAuthDB()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	exists, err := db.SessionExistsByUsername("giorno-giovanna")
	if err != nil {
		t.Error(err)
	}
	if !exists {
		t.Errorf("got %v, want %v", exists, true)
	}
	exists, err = db.SessionExists("bruno-bucciarati")
	if err != nil {
		t.Error(err)
	}
	if exists {
		t.Errorf("got %v, want %v", exists, false)
	}
}

func TestCheckPassword(t *testing.T) {
	db, err := initAuthDB()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	match, err := db.CheckPassword("giorno-giovanna", "goldenwind")
	if err != nil {
		t.Error(err)
	}
	if !match {
		t.Errorf("got %v, want %v", match, true)
	}
	match, err = db.CheckPassword("giorno-giovanna", "goldenwindrequiem")
	if err != nil {
		t.Error(err)
	}
	if match {
		t.Errorf("got %v, want %v", match, false)
	}
}

func TestCheckSessionTokenExpiry(t *testing.T) {
	db, err := initAuthDB()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	expired, err := db.CheckSessionTokenExpiry("goldenexperience")
	if err != nil {
		t.Error(err)
	}
	if !expired {
		t.Errorf("got %v, want %v", expired, false)
	}
	expired, err = db.CheckSessionTokenExpiry("horahorahora")
	if err != nil {
		t.Error(err)
	}
	if expired {
		t.Errorf("got %v, want %v", expired, true)
	}
}

func TestRefreshSessionToken(t *testing.T) {
	db, err := initAuthDB()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	n, err := db.RefreshSessionToken("goldenexperience", "silverexperience", time.Now().Add(1*time.Minute))
	if err != nil {
		t.Error(err)
	}
	if n != 1 {
		t.Errorf("got %d, want %d", n, 1)
	}
	n, err = db.RefreshSessionToken("goldenexperience", "silverexperience", time.Now().Add(1*time.Minute))
	if err != nil {
		t.Error(err)
	}
	if n != 0 {
		t.Errorf("got %d, want %d", n, 0)
	}
}

func TestDeleteSessionToken(t *testing.T) {
	db, err := initAuthDB()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	n, err := db.DeleteSessionToken("goldenexperience")
	if err != nil {
		t.Error(err)
	}
	if n != 1 {
		t.Errorf("got %v, want %v", n, 1)
	}
	n, err = db.DeleteSessionToken("goldenexperience")
	if err != nil {
		t.Error(err)
	}
	if n != 0 {
		t.Errorf("got %v, want %v", n, 0)
	}
}

func TestDeleteSessionTokenByUsername(t *testing.T) {
	db, err := initAuthDB()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	n, err := db.DeleteSessionTokenByUsername("giorno-giovanna")
	if err != nil {
		t.Error(err)
	}
	if n != 1 {
		t.Errorf("got %v, want %v", n, 1)
	}
	n, err = db.DeleteSessionTokenByUsername("bruno-bucciarati")
	if err != nil {
		t.Error(err)
	}
	if n != 0 {
		t.Errorf("got %v, want %v", n, 0)
	}
}
