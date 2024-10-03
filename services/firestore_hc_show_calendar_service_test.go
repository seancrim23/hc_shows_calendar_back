package services

import (
	"context"
	"fmt"
	"hc_shows_backend/models"
	"hc_shows_backend/utils"
	"log"
	"os"
	"os/exec"
	"testing"
	"time"
)

var testUser1 = models.User{
	Username: "coolpromoter123",
	Email:    "coolpromoter123@hotmail.com",
	Hash:     "reallycoolpassword45",
}

var testUser2 = models.User{
	Username: "coolpromoter456",
	Email:    "coolpromoter456@hotmail.com",
	Hash:     "reallycoolpassword45",
}

var testShow1 = models.Show{
	Id:       "abc123",
	Lineup:   models.Lineup{"cool band 1", "cool band 2", "cool band 3", "cool band 4"},
	Date:     time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC),
	State:    "MD",
	City:     "Baltimore",
	Venue:    "Charlie's Chowder Chapel",
	Address:  "123 Chowder Avenue",
	Promoter: "coolpromoter123",
}

var testShow2 = models.Show{
	Id:       "abc456",
	Lineup:   models.Lineup{"cool band 1", "cool band 2", "cool band 3", "cool band 4"},
	Date:     time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC),
	State:    "RI",
	City:     "Pawtucket",
	Venue:    "Hal's Hot Dog Hut",
	Address:  "123 Burger Ave",
	Promoter: "coolpromoter456",
}

var testShow3 = models.Show{
	Id:       "xyz123",
	Lineup:   models.Lineup{"cool band 1", "cool band 2", "cool band 3", "cool band 4"},
	Date:     time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC),
	State:    "RI",
	City:     "Quahog",
	Venue:    "The Drunken Clam",
	Address:  "123 The Clam Ave",
	Promoter: "coolpromoter456",
}

// better way to do this would be start the firestore emulator command
// wait until it is set up
// run the tests
// clean up
func TestMain(m *testing.M) {
	os.Setenv(utils.FIRESTORE_EMULATOR_HOST, utils.LOCALHOST+":"+utils.FIRESTORE_EMULATOR_PORT)
	os.Setenv(utils.GCP_PROJECT_ID, "hc-show-calendar")
	go setupTests()
	//sleep to ensure the emulator starts...
	time.Sleep(5 * time.Second)
	loadTestData()
	code := m.Run()
	cleanupTests()
	os.Exit(code)
}

func setupTests() {
	fmt.Println("setting up tests")
	cmd := exec.CommandContext(context.Background(), "gcloud", "emulators", "firestore", "start", "--host-port="+utils.LOCALHOST+":"+utils.FIRESTORE_EMULATOR_PORT)
	//this command blocks... maybe make this a go routine too?
	//look into if there's any way i can make it not block
	err := cmd.Run()
	if err != nil {
		log.Fatal(err)
	}
}

func loadTestData() {
	fmt.Println("loading test data...")
	service, close, err := NewFirestoreHCShowCalendarService()
	defer close()
	if err != nil {
		fmt.Println("failure setting up test data...")
		log.Fatal(err)
	}

	_, err = service.CreateUser(testUser1)
	if err != nil {
		fmt.Println("failure setting up test user data...")
		log.Fatal(err)
	}
	_, err = service.CreateUser(testUser2)
	if err != nil {
		fmt.Println("failure setting up test user data...")
		log.Fatal(err)
	}
	_, err = service.CreateShow(testShow1)
	if err != nil {
		fmt.Println("failure setting up test show data...")
		log.Fatal(err)
	}
	_, err = service.CreateShow(testShow2)
	if err != nil {
		fmt.Println("failure setting up test show data...")
		log.Fatal(err)
	}
	_, err = service.CreateShow(testShow1)
	if err != nil {
		fmt.Println("failure setting up test show data...")
		log.Fatal(err)
	}
	_, err = service.CreateShow(testShow3)
	if err != nil {
		fmt.Println("failure setting up test show data...")
		log.Fatal(err)
	}
}

func cleanupTests() {
	fmt.Println("cleaning up tests")
	cmd := exec.Command("curl", "-d", "''", utils.LOCALHOST+":"+utils.FIRESTORE_EMULATOR_PORT+"/shutdown")
	if err := cmd.Run(); err != nil {
		log.Fatal(err)
	}
}

func TestGetShows(t *testing.T) {
	service, close, err := NewFirestoreHCShowCalendarService()
	defer close()
	if err != nil {
		fmt.Println("failure starting service...")
		log.Fatal(err)
	}

	t.Run("can get all shows with no query filters passed in", func(t *testing.T) {
		testEmptyFilters := map[string]string{}
		shows, err := service.GetShows(testEmptyFilters)
		if err != nil {
			fmt.Println("failure getting test show data...")
			log.Fatal(err)
		}
		if len(*shows) != 3 {
			t.Fatalf("expected 3 shows in response but got %d", len(*shows))
		}
	})
	t.Run("can get correct show when one query filter passed in", func(t *testing.T) {
		testQueryFilter := map[string]string{
			"city": "Baltimore",
		}
		shows, err := service.GetShows(testQueryFilter)
		if err != nil {
			fmt.Println("failure getting test show data...")
			log.Fatal(err)
		}

		if len(*shows) != 1 {
			t.Fatalf("expected 1 shows in response but got %d", len(*shows))
		}
		if (*shows)[0].City != "Baltimore" {
			t.Fatalf("expected city of %q in response", (*shows)[0].City)
		}
	})
	t.Run("can get correct show when two query filters are passed in", func(t *testing.T) {
		testQueryFilter := map[string]string{
			"city":  "Quahog",
			"state": "RI",
		}
		shows, err := service.GetShows(testQueryFilter)
		if err != nil {
			fmt.Println("failure getting test show data...")
			log.Fatal(err)
		}

		if len(*shows) != 1 {
			t.Fatalf("expected 1 shows in response but got %d", len(*shows))
		}
		if (*shows)[0].City != "Quahog" {
			t.Fatalf("expected city of %q in response got %q", "Quahog", (*shows)[0].City)
		}
		if (*shows)[0].State != "RI" {
			t.Fatalf("expected state of %q in response got %q", "RI", (*shows)[0].State)
		}
	})
}

func TestGetShow(t *testing.T) {
	service, close, err := NewFirestoreHCShowCalendarService()
	defer close()
	if err != nil {
		fmt.Println("failure starting service...")
		log.Fatal(err)
	}

	t.Run("can get a show with the correct show id", func(t *testing.T) {
		show, err := service.GetShow(testShow1.Id)
		if err != nil {
			fmt.Println("failure getting test show data...")
			log.Fatal(err)
		}
		if show.Id != testShow1.Id {
			t.Fatalf("expected id of %q in response but got %q", show.Id, testShow1.Id)
		}
	})
	t.Run("does not get a show with an invalid show id", func(t *testing.T) {
		_, err := service.GetShow("badshowid")
		if err == nil {
			t.Fatalf("expected error when bad show id is passed to service")
		}
	})
}

func TestCreateShow(t *testing.T) {
	service, close, err := NewFirestoreHCShowCalendarService()
	defer close()
	if err != nil {
		fmt.Println("failure starting service...")
		log.Fatal(err)
	}

	t.Run("can create a show with valid show info", func(t *testing.T) {
		var testShow = models.Show{
			Id:       "aaa111",
			Lineup:   models.Lineup{"cool band 1", "cool band 2", "cool band 3", "cool band 4"},
			Date:     time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC),
			State:    "RI",
			City:     "Quahog",
			Venue:    "The Drunken Clam",
			Address:  "123 The Clam Ave",
			Promoter: "coolpromoter456",
		}
		show, err := service.CreateShow(testShow)
		if err != nil {
			fmt.Println("failure creating test show data...")
			log.Fatal(err)
		}
		if show.Id != testShow.Id {
			t.Fatalf("expected id of %q in response but got %q", show.Id, testShow.Id)
		}
		err = service.DeleteShow(testShow.Id)
		if err != nil {
			fmt.Println("failure deleting test show data...")
			log.Fatal(err)
		}
	})
}

func TestUpdateShow(t *testing.T) {
	service, close, err := NewFirestoreHCShowCalendarService()
	defer close()
	if err != nil {
		fmt.Println("failure starting service...")
		log.Fatal(err)
	}

	t.Run("can correctly update a show", func(t *testing.T) {
		var testShow = models.Show{
			Id:       "xyz123",
			Lineup:   models.Lineup{"cool band 1", "cool band 2", "cool band 3", "cool band 4"},
			Date:     time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC),
			State:    "RI",
			City:     "Quahog",
			Venue:    "Chris' Corn Castle",
			Address:  "123 Corn St",
			Promoter: "coolpromoter456",
		}

		show, err := service.UpdateShow(testShow.Id, testShow)
		if err != nil {
			fmt.Println("failure updating show data")
			log.Fatal(err)
		}

		if (*show).Venue != testShow.Venue {
			t.Fatalf("expected venue of %q in response got %q", testShow.Venue, (*show).Venue)
		}
		if (*show).Address != testShow.Address {
			t.Fatalf("expected address of %q in response got %q", testShow.Address, (*show).Address)
		}

	})
}

func TestDeleteShow(t *testing.T) {
	service, close, err := NewFirestoreHCShowCalendarService()
	defer close()
	if err != nil {
		fmt.Println("failure starting service...")
		log.Fatal(err)
	}

	t.Run("can create a show and then delete a show with valid show info", func(t *testing.T) {
		var testShow = models.Show{
			Id:       "aaa111",
			Lineup:   models.Lineup{"cool band 1", "cool band 2", "cool band 3", "cool band 4"},
			Date:     time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC),
			State:    "RI",
			City:     "Quahog",
			Venue:    "The Drunken Clam",
			Address:  "123 The Clam Ave",
			Promoter: "coolpromoter456",
		}
		_, err := service.CreateShow(testShow)
		if err != nil {
			fmt.Println("failure creating test show data...")
			log.Fatal(err)
		}

		err = service.DeleteShow(testShow.Id)
		if err != nil {
			t.Fatalf("expected successful delete of show")
		}
	})
}

func TestGetUser(t *testing.T) {
	service, close, err := NewFirestoreHCShowCalendarService()
	defer close()
	if err != nil {
		fmt.Println("failure starting service...")
		log.Fatal(err)
	}

	t.Run("can get a user with the correct user id", func(t *testing.T) {
		user, err := service.GetUser(testUser1.Username)
		if err != nil {
			fmt.Println("failure getting test user data...")
			log.Fatal(err)
		}
		if user.Username != testUser1.Username {
			t.Fatalf("expected id of %q in response but got %q", testUser1.Username, user.Username)
		}
	})
	t.Run("does not get a user with an invalid user id", func(t *testing.T) {
		_, err := service.GetUser("baduserid")
		if err == nil {
			t.Fatalf("expected error when bad user id is passed to service")
		}
	})
}

func TestCreateUser(t *testing.T) {
	service, close, err := NewFirestoreHCShowCalendarService()
	defer close()
	if err != nil {
		fmt.Println("failure starting service...")
		log.Fatal(err)
	}

	t.Run("can create a user with valid user info", func(t *testing.T) {
		var testUser = models.User{
			Username: "coolpromoter111",
			Email:    "coolpromoter111@hotmail.com",
			Hash:     "reallycoolpassword45",
		}
		user, err := service.CreateUser(testUser)
		if err != nil {
			fmt.Println("failure creating test user data...")
			log.Fatal(err)
		}
		if user.Username != testUser.Username {
			t.Fatalf("expected id of %q in response but got %q", user.Username, testUser.Username)
		}
		err = service.DeleteUser(testUser.Username)
		if err != nil {
			fmt.Println("failure deleting test user data...")
			log.Fatal(err)
		}
	})
}

func TestUpdateUser(t *testing.T) {
	service, close, err := NewFirestoreHCShowCalendarService()
	defer close()
	if err != nil {
		fmt.Println("failure starting service...")
		log.Fatal(err)
	}

	t.Run("can correctly update a user", func(t *testing.T) {
		var testUser = models.User{
			Username: "coolupdatedemail",
			Email:    "coolupdatedemail@hotmail.com",
		}

		user, err := service.UpdateUser(testUser.Username, testUser)
		if err != nil {
			fmt.Println("failure updating user data")
			log.Fatal(err)
		}

		if (*user).Email != testUser.Email {
			t.Fatalf("expected email of %q in response got %q", testUser.Email, (*user).Email)
		}
	})
}

func TestDeleteUser(t *testing.T) {
	service, close, err := NewFirestoreHCShowCalendarService()
	defer close()
	if err != nil {
		fmt.Println("failure starting service...")
		log.Fatal(err)
	}

	t.Run("can create a user and successfully delete with valid user info", func(t *testing.T) {
		var testUser = models.User{
			Username: "coolpromoter111",
			Email:    "coolpromoter111@hotmail.com",
			Hash:     "reallycoolpassword45",
		}
		user, err := service.CreateUser(testUser)
		if err != nil {
			fmt.Println("failure creating test user data...")
			log.Fatal(err)
		}
		err = service.DeleteUser(user.Username)
		if err != nil {
			t.Fatalf("expected successful delete of user")
		}
	})
}

func TestAuthUser(t *testing.T) {
	service, close, err := NewFirestoreHCShowCalendarService()
	defer close()
	if err != nil {
		fmt.Println("failure starting service...")
		log.Fatal(err)
	}

	t.Run("can create a user and successfully authorize the user with valid user info", func(t *testing.T) {
		var testUser = models.User{
			Username: "coolpromoter111",
			Email:    "coolpromoter111@hotmail.com",
			Pass:     "reallycoolpassword45",
		}
		_, err := service.CreateUser(testUser)
		if err != nil {
			fmt.Println("failure creating test user data...")
			log.Fatal(err)
		}
		token, err := service.AuthUser(testUser)
		if err != nil {
			fmt.Println("failure authorizing test user ...")
			log.Fatal(err)
		}
		if token == "" {
			t.Fatalf("expected a token from successful user auth")
		}
	})
}
