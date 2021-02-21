package proj2

// You MUST NOT change what you import.  If you add ANY additional
// imports it will break the autograder, and we will be Very Upset.

import (
	"testing"
	"reflect"
	"github.com/cs161-staff/userlib"
	"encoding/json"
	_ "encoding/hex"
	_ "github.com/google/uuid"
	"strings"
	_ "errors"
	_ "strconv"
)

func clear() {
	// Wipes the storage so one test does not affect another
	userlib.DatastoreClear()
	userlib.KeystoreClear()
}

func TestRegister(t *testing.T) {
	clear()

	u, err := InitUser("", "fubar")
	if u != nil {
		// t.Error says the test fails
		t.Error("Failed to check empty username", err)
		return
	}

	p, err := InitUser("alice", "")
	if p != nil {
		// t.Error says the test fails
		t.Error("Failed to check empty password", err)
		return
	}
}

func TestInitAndGet(t *testing.T) {
	clear()
	t.Log("Initialization test")

	// You can set this to false!
	userlib.SetDebugStatus(true)

	print("Creating Alice")
	alice, err := InitUser("alice", "fubar")
	if err != nil {
		// t.Error says the test fails
		t.Error("Failed to initialize user", err)
		return
	}
	print("Created Alice")
	// t.Log() only produces output if you run with "go test -v"
	t.Log("Got user", alice)
	// If you want to comment the line above,
	// write _ = u here to make the compiler happy
	// You probably want many more tests here.
	getAlice, err := GetUser("alice", "fubar")
	print("got Alice")
	if getAlice == nil || err != nil {
		t.Error(err)
		return
	}
	print("success got alice")
	aliceBytes, _  := json.Marshal(alice)
	getAliceBytes, _ := json.Marshal(getAlice)
	if !reflect.DeepEqual(aliceBytes, getAliceBytes) {
		t.Error("Init and Get userdata are not the same.")
		return
	}

	_, err = GetUser("alice", "incorrect")
	if err == nil {
		t.Error("Password wrong but still got user")
		return
	}

	_, err = GetUser("incorrect", "fubar")
	if err == nil {
		t.Error("Username wrong but still got user")
		return
	}

	getAlice2, err := GetUser("alice", "fubar")
	aliceBytes2, _  := json.Marshal(alice)
	getAliceBytes2, _ := json.Marshal(getAlice2)
	if !reflect.DeepEqual(aliceBytes2, getAliceBytes2) {
		t.Error("Init and Get userdata are not the same.")
		return
	}
}

func TestInitAndGetUsingMalDatastore(t *testing.T) {
	userlib.SetDebugStatus(true)
	userlib.DatastoreClear()
	userlib.KeystoreClear()
	datastore := userlib.DatastoreGetMap()
	keystore := userlib.KeystoreGetMap()
	_, _ = datastore, keystore

	_, err := InitUser("bob", "fubar")
	if err != nil {
		// t.Error says the test fails
		t.Error("Failed to initialize user", err)
		return
	}

	_, err = InitUser("alice", "fubar")
	if err != nil {
		// t.Error says the test fails
		t.Error("Failed to initialize user", err)
		return
	}

	var userKeys []userlib.UUID
	var values [][]byte

	for key, value := range datastore {
		userKeys = append(userKeys, key)
		values = append(values, value)
	}

	userlib.DatastoreSet(userKeys[0], values[1])

	i := 1
	for i < len(userKeys) {
		userlib.DatastoreSet(userKeys[i], values[0])
		i += 1
	}

	_, err = GetUser("bob", "fubar")
	if err == nil {
		t.Error("Shouldn't have been able to retrive Bob", err)
		return
	}

	_, err = GetUser("alice", "fubar")
	if err == nil {
		t.Error("Shouldn't have been able to retrive Alice", err)
		return
	}
}

func TestInitAndGetUsingEmptyDataStore(t *testing.T) {

	userlib.DatastoreClear()
	userlib.KeystoreClear()

	_, err := InitUser("bob", "fubar")
	if err != nil {
		// t.Error says the test fails
		t.Error("Failed to initialize user", err)
		return
	}

	userlib.DatastoreClear()

	_, err = GetUser("bob", "fubar")
	if err == nil {
		t.Error("No data in DataStore but still retrieved user Bob")
		return
	}

	userlib.DatastoreClear()
	userlib.KeystoreClear()

	datastore := userlib.DatastoreGetMap()

	_, err = InitUser("bob", "fubar")
	if err != nil {
		// t.Error says the test fails
		t.Error("Failed to initialize user", err)
		return
	}

	_, err = InitUser("alice", "fubar")
	if err != nil {
		// t.Error says the test fails
		t.Error("Failed to initialize user", err)
		return
	}


	var userKeys []userlib.UUID
	var values [][]byte
	for key, value := range datastore {
		userKeys = append(userKeys, key)
		values = append(values, value)
	}
	datastore[userKeys[0]] = userlib.RandomBytes(len(userKeys[0]))

	_, err0 := GetUser("bob", "fubar")
	_, err1 := GetUser("alice", "fubar")

	if err0 == nil && err1 == nil {
		t.Error("Should not have been able to retrieve Alice and Bob")
	}
}

func TestStorage(t *testing.T) {
	clear()
	u, err := InitUser("alice", "fubar")
	if err != nil {
		t.Error("Failed to initialize user", err)
		return
	}

	v := []byte("This is a test")
	u.StoreFile("file1", v)

	v2, err2 := u.LoadFile("file1")
	if err2 != nil {
		t.Error("Failed to upload and download", err2)
		return
	}
	if !reflect.DeepEqual(v, v2) {
		t.Error("Downloaded file is not the same", v, v2)
		return
	}

	overwrite := []byte("This is still a test")
	u.StoreFile("file1", overwrite)
	overwriteGet, err := u.LoadFile("file1")
	if err != nil {
		t.Error("Failed to overwrite exisiting file or retrieve overwritten file", err)
		return
	}
	if !reflect.DeepEqual(overwrite, overwriteGet) {
		t.Error("Downloaded file is not the same", v, v2)
		return
	}
}

func TestEmptyFilenameStorage(t *testing.T) {
	clear()
	u, err := InitUser("alice", "fubar")
	if err != nil {
		t.Error("Failed to initialize user", err)
		return
	}

	v := []byte("test")
	u.StoreFile("", v)

	v2, err2 := u.LoadFile("")
	if err2 != nil {
		t.Error("Failed to upload and download", err2)
		return
	}
	if !reflect.DeepEqual(v, v2) {
		t.Error("Downloaded file is not the same", v, v2)
		return
	}
}

func TestStorageUsingMalDatastore(t *testing.T) {
	userlib.SetDebugStatus(true)
	userlib.DatastoreClear()
	userlib.KeystoreClear()

	datastore := userlib.DatastoreGetMap()
	userlib.KeystoreGetMap()

	files := []string{"file1", "file2", "file3", "file4", "file5"}
	users := []string{"user1", "user2", "user3", "user4", "user5"}

	for i, offset := range []int{-12, -5, 0, 3, 15 } {
		u, _ := InitUser(users[i], "pass")

		file := userlib.RandomBytes(userlib.AESBlockSize * (i + 1) - offset)
		u.StoreFile(files[i], file)

		u, _ = GetUser(users[i], "pass")

		var userKeys []userlib.UUID
		var values [][]byte
		for key, value := range datastore {
			userKeys = append(userKeys, key)
			values = append(values, value)
		}

		didError := false

		for key := range userKeys {
			datastore[userKeys[key]] = userlib.RandomBytes(len(values[key]))
			_, err := u.LoadFile(files[i])
			if err != nil {
				// Should error here
				didError = true
			}
			datastore[userKeys[key]] = values[key]
		}

		if !didError {
			t.Error("Should not have been able to load file")
		}
	}
}

func TestInvalidFile(t *testing.T) {
	clear()
	u, err := InitUser("alice", "fubar")
	if err != nil {
		t.Error("Failed to initialize user", err)
		return
	}

	_, err2 := u.LoadFile("this file does not exist")
	if err2 == nil {
		t.Error("Downloaded a ninexistent file", err2)
		return
	}
}

func TestLoadBad(t *testing.T) {
	clear()
	u, err := InitUser("alice", "fubar")
	if err != nil {
		t.Error("Failed to initialize user", err)
		return
	}
	_, err = u.LoadFile("file1")
	if err == nil {
		t.Error("Loaded a file that does not exist", err)
		return
	}
}

func TestLoadImmediately(t *testing.T) {
	clear()
	_, err := InitUser("alice", "fubar")
	if err != nil {
		t.Error("Failed to initialize user", err)
		return
	}

	v := []byte("This is a test")
	getAlice, err := GetUser("alice", "fubar")
	getAlice2, err := GetUser("alice", "fubar")

	getAlice.StoreFile("file1", v)
	get1, err := getAlice.LoadFile("file1")
	if  err != nil {
		t.Error("Can't load file from getAlice", err)
		return
	}
	if !reflect.DeepEqual(v, get1) {
		t.Error("Downloaded file is not the same", v, get1)
		return
	}

	get2, err := getAlice2.LoadFile("file1")
	if err != nil {
		t.Error("Could not load file", err)
		return
	}
	if !reflect.DeepEqual(get1, get2) {
		t.Error("Downloaded file is not the same", get1, get2)
		return
	}
	if !reflect.DeepEqual(v, get2) {
		t.Error("Downloaded file is not the same", v, get2)
		return
	}

	v2 := []byte("End of File")
	v2Final := []byte("This is a testEnd of File")
	getAlice.AppendFile("file1", v2)

	get1, err = getAlice.LoadFile("file1")
	if  err != nil {
		t.Error("Can't load file from getAlice", err)
		return
	}
	if !reflect.DeepEqual(v2Final, get1) {
		t.Error("Downloaded file is not the same", v2Final, get1)
		return
	}

	get2, err = getAlice2.LoadFile("file1")
	if err != nil {
		t.Error("Could not load file", err)
		return
	}
	if !reflect.DeepEqual(get1, get2) {
		t.Error("Downloaded file is not the same", get1, get2)
		return
	}
	if !reflect.DeepEqual(v2Final, get2) {
		t.Error("Downloaded file is not the same", v2Final, get2)
		return
	}
}

func TestMultipleInstances(t *testing.T) {
	clear()
	_, err := InitUser("alice", "fubar")
	if err != nil {
		t.Error("Failed to initialize user", err)
		return
	}
	u, err := InitUser("bob", "foobar")
	if err != nil {
		t.Error("Failed to initialize user", err)
		return
	}

	v := []byte("This is a test")
	u.StoreFile("file1", v)

	getAlice, err := GetUser("alice", "fubar")
	getAlice2, err := GetUser("alice", "fubar")

	magic_string, err := u.ShareFile("file1", "alice")
	err = getAlice.ReceiveFile("fileAlice", "bob", magic_string)

	v2 := []byte("End of File")
	v2Final := []byte("This is a testEnd of File")

	getAlice2.AppendFile("fileAlice", v2)	
	loaded, err := getAlice.LoadFile("fileAlice")
	if err != nil {
		t.Error("Could not load file", err)
		return
	}

	if !reflect.DeepEqual(v2Final, loaded) {
		t.Error("Downloaded file is not the same", v2Final, loaded)
		return
	}

	v3 := []byte("This is a test3")
	u.StoreFile("file3", v3)
	magic_string, err = u.ShareFile("file3", "alice")
	err = getAlice.ReceiveFile("fileAlice3", "bob", magic_string)
	loaded3, err := getAlice2.LoadFile("fileAlice3")

	if !reflect.DeepEqual(v3, loaded3) {
		t.Error("Downloaded file is not the same", v3, loaded3)
		return
	}
	loadedAlice3, err := getAlice.LoadFile("fileAlice3")
	if !reflect.DeepEqual(v3, loadedAlice3) {
		t.Error("Downloaded file is not the same", v3, loadedAlice3)
		return
	}

}


func TestAppendBasic(t *testing.T) {
	userlib.SetDebugStatus(false)
	userlib.DatastoreClear()
	userlib.KeystoreClear()
	datastore := userlib.DatastoreGetMap()
	keystore := userlib.KeystoreGetMap()
	_, _ = datastore, keystore

	user, _ := InitUser("Bob", "fubar")
	fileContent := []byte("Start of File")
	fileContentAppend := []byte("End of File")
	fileContentFinal := []byte("Start of FileEnd of File")
	user.StoreFile("file1", fileContent)
	user.AppendFile("file1", fileContentAppend)
	loadedFile, _ := user.LoadFile("file1")
	if !reflect.DeepEqual(fileContentFinal, loadedFile) {
				t.Error("Appended file is not the same as reference\n",
					fileContentFinal, "\n", loadedFile)
				return
	}
}

func TestAppendComplex(t *testing.T) {
	userlib.SetDebugStatus(false)
	userlib.DatastoreClear()
	userlib.KeystoreClear()
	datastore := userlib.DatastoreGetMap()
	keystore := userlib.KeystoreGetMap()
	_, _ = datastore, keystore

	user, _ := InitUser("Bob", "fubar")
	fileContent := []byte("Start of File")
	fileContentAppend := []byte("End of File")
	fileContentFinal := []byte("Start of FileEnd of File")
	user.StoreFile("file1", fileContent)
	user.AppendFile("file1", fileContentAppend)

	u2, err := InitUser("alice", "fubar")
	if err != nil {
		t.Error("Failed to initialize user", err)
		return
	}

	magic_string, err := user.ShareFile("file1", "alice")
	if err != nil {
		t.Error("Failed to share the a file", err)
		return
	}
	err = u2.ReceiveFile("file2", "Bob", magic_string)
	if err != nil {
		t.Error("Failed to receive the share message", err)
		return
	}

	loadedFile, _ := user.LoadFile("file1")
	if !reflect.DeepEqual(fileContentFinal, loadedFile) {
				t.Error("Appended file is not the same as reference\n",
					fileContentFinal, "\n", loadedFile)
				return
	}

	loadedFile, _ = u2.LoadFile("file2")
	if !reflect.DeepEqual(fileContentFinal, loadedFile) {
				t.Error("Appended file is not the same as reference\n",
					fileContentFinal, "\n", loadedFile)
				return
	}

	u2.AppendFile("file2", fileContentAppend)
	user.AppendFile("file1", fileContentAppend)
	fileContentFinal = []byte("Start of FileEnd of FileEnd of FileEnd of File")

	loadedFile, _ = user.LoadFile("file1")
	if !reflect.DeepEqual(fileContentFinal, loadedFile) {
				t.Error("Appended file is not the same as reference\n",
					fileContentFinal, "\n", loadedFile)
				return
	}

	loadedFile, _ = u2.LoadFile("file2")
	if !reflect.DeepEqual(fileContentFinal, loadedFile) {
				t.Error("Appended file is not the same as reference\n",
					fileContentFinal, "\n", loadedFile)
				return
	}

}

func TestAppendLong(t *testing.T) {
	clear()
	u, err := InitUser("alice", "fubar")
	if err != nil {
		t.Error("Failed to init user", err)
		return
	}

	v := []byte(strings.Repeat("We are using something very long. ", 20))
	u.StoreFile("file1", v)

	v2 := []byte(strings.Repeat("We are appending something very long. ", 10))
	err = u.AppendFile("file1", v2)
	if err != nil {
		t.Error("Failed to append file", err)
		return
	}

	v3, err := u.LoadFile("file1")
	if err != nil {
		t.Error("Failed to upload and download", err)
	}
	if !reflect.DeepEqual(append(v, v2...), v3) {
		t.Error("Downloaded appended file is not the same", v, v2)
	}
}

func TestAppendMultipleShares(t *testing.T) {
	userlib.SetDebugStatus(true)
	userlib.DatastoreClear()
	userlib.KeystoreClear()
	datastore := userlib.DatastoreGetMap()
	keystore := userlib.KeystoreGetMap()
	_, _ = datastore, keystore

	user1, _ := InitUser("user1", "pass")
	user2, _ := InitUser("user2", "pass")
	user3, _ := InitUser("user3", "pass")
	user4, _ := InitUser("user4", "pass")

	fileContent := []byte("Start of File")
	user1.StoreFile("file1", fileContent)
	magic_string, err := user1.ShareFile("file1", "user2")
	if err != nil {
		t.Error("Failed to share the a file", err)
		return
	}
	err = user2.ReceiveFile("file2", "user1", magic_string)
	if err != nil {
		t.Error("Failed to receive the share message", err)
		return
	}
	magic_string, err = user2.ShareFile("file2", "user3")
	if err != nil {
		t.Error("Failed to share the a file", err)
		return
	}
	err = user3.ReceiveFile("file3", "user2", magic_string)
	if err != nil {
		t.Error("Failed to receive the share message", err)
		return
	}
	magic_string, err = user1.ShareFile("file1", "user4")
	if err != nil {
		t.Error("Failed to share the a file", err)
		return
	}
	err = user4.ReceiveFile("file4", "user1", magic_string)
	if err != nil {
		t.Error("Failed to receive the share message", err)
		return
	}


	emptyContent := []byte("")
	appendContent := []byte("Middle of File")
	resultingContent := []byte("Start of FileMiddle of File")
	user1.AppendFile("file1", appendContent)
	user1.AppendFile("file1", emptyContent)

	contentLoaded, _ := user1.LoadFile("file1")
	if !reflect.DeepEqual(resultingContent, contentLoaded) {
		t.Error("Shared file is not the same", resultingContent, contentLoaded)
		return
	}

	contentLoaded, _ = user2.LoadFile("file2")
	if !reflect.DeepEqual(resultingContent, contentLoaded) {
		t.Error("Shared file is not the same", resultingContent, contentLoaded)
		return
	}

	contentLoaded, _ = user3.LoadFile("file3")
	if !reflect.DeepEqual(resultingContent, contentLoaded) {
		t.Error("Shared file is not the same", resultingContent, contentLoaded)
		return
	}

	contentLoaded, _ = user4.LoadFile("file4")
	if !reflect.DeepEqual(resultingContent, contentLoaded) {
		t.Error("Shared file is not the same", resultingContent, contentLoaded)
		return
	}


	// User 3 append to file. U1, U2, U3, and U4 should all see update
	user3.AppendFile("file3", appendContent)
	resultingContent = []byte("Start of FileMiddle of FileMiddle of File")

	contentLoaded, _ = user1.LoadFile("file1")
	if !reflect.DeepEqual(resultingContent, contentLoaded) {
		t.Error("Shared file is not the same", resultingContent, contentLoaded)
		return
	}

	contentLoaded, _ = user2.LoadFile("file2")
	if !reflect.DeepEqual(resultingContent, contentLoaded) {
		t.Error("Shared file is not the same", resultingContent, contentLoaded)
		return
	}

	contentLoaded, _ = user3.LoadFile("file3")
	if !reflect.DeepEqual(resultingContent, contentLoaded) {
		t.Error("Shared file is not the same", resultingContent, contentLoaded)
		return
	}

	contentLoaded, _ = user4.LoadFile("file4")
	if !reflect.DeepEqual(resultingContent, contentLoaded) {
		t.Error("Shared file is not the same", resultingContent, contentLoaded)
		return
	}

	// User 4 append to file. U1, U2, U3, and U4 should all see update
	user4.AppendFile("file4", appendContent)
	resultingContent = []byte("Start of FileMiddle of FileMiddle of FileMiddle of File")

	contentLoaded, _ = user1.LoadFile("file1")
	if !reflect.DeepEqual(resultingContent, contentLoaded) {
		t.Error("Shared file is not the same", resultingContent, contentLoaded)
		return
	}

	contentLoaded, _ = user2.LoadFile("file2")
	if !reflect.DeepEqual(resultingContent, contentLoaded) {
		t.Error("Shared file is not the same", resultingContent, contentLoaded)
		return
	}

	contentLoaded, _ = user3.LoadFile("file3")
	if !reflect.DeepEqual(resultingContent, contentLoaded) {
		t.Error("Shared file is not the same", resultingContent, contentLoaded)
		return
	}

	contentLoaded, _ = user4.LoadFile("file4")
	if !reflect.DeepEqual(resultingContent, contentLoaded) {
		t.Error("Shared file is not the same", resultingContent, contentLoaded)
		return
	}

	// User 1 append to file. U1, U2, U3, and U4 should all see update
	user1.AppendFile("file1", appendContent)
	resultingContent = []byte("Start of FileMiddle of FileMiddle of FileMiddle of FileMiddle of File")

	contentLoaded, _ = user1.LoadFile("file1")
	if !reflect.DeepEqual(resultingContent, contentLoaded) {
		t.Error("Shared file is not the same", resultingContent, contentLoaded)
		return
	}

	contentLoaded, _ = user2.LoadFile("file2")
	if !reflect.DeepEqual(resultingContent, contentLoaded) {
		t.Error("Shared file is not the same", resultingContent, contentLoaded)
		return
	}

	contentLoaded, _ = user3.LoadFile("file3")
	if !reflect.DeepEqual(resultingContent, contentLoaded) {
		t.Error("Shared file is not the same", resultingContent, contentLoaded)
		return
	}

	contentLoaded, _ = user4.LoadFile("file4")
	if !reflect.DeepEqual(resultingContent, contentLoaded) {
		t.Error("Shared file is not the same", resultingContent, contentLoaded)
		return
	}

	// Revoke access from User3
	err = user1.RevokeFile("file1", "user2")
	if err != nil {
		t.Error("Could not revoke file")
		return
	}

	// User 2 append to file. U1 and U4 should all see update
	user4.AppendFile("file4", appendContent)
	resultingContent = []byte("Start of FileMiddle of FileMiddle of FileMiddle of FileMiddle of FileMiddle of File")

	contentLoaded, _ = user1.LoadFile("file1")
	if !reflect.DeepEqual(resultingContent, contentLoaded) {
		t.Error("Shared file is not the same", resultingContent, contentLoaded)
		return
	}

	contentLoaded, _ = user2.LoadFile("file2")
	if reflect.DeepEqual(resultingContent, contentLoaded) {
		t.Error("User2 should not be able to see the updated file", resultingContent, contentLoaded)
		return
	}

	contentLoaded, err = user3.LoadFile("file3")
	if reflect.DeepEqual(resultingContent, contentLoaded) {
		t.Error("User3 should not be able to see the updated file", resultingContent, contentLoaded)
		return
	}

	contentLoaded, _ = user4.LoadFile("file4")
	if !reflect.DeepEqual(resultingContent, contentLoaded) {
		t.Error("Shared file is not the same", resultingContent, contentLoaded)
		return
	}
}

func TestAppendWithCorruptDatastore(t *testing.T) {
	userlib.SetDebugStatus(true)
	userlib.DatastoreClear()
	userlib.KeystoreClear()
	datastore := userlib.DatastoreGetMap()
	keystore := userlib.KeystoreGetMap()
	_, _ = datastore, keystore

	u1, err := InitUser("user1", "pass")
	if err != nil {
		t.Error(err)
		return
	}

	fileContents := userlib.RandomBytes(userlib.AESBlockSize)
	appendContents := userlib.RandomBytes(userlib.AESBlockSize)

	u1.StoreFile("file1", fileContents)

	var userKeys []userlib.UUID
	var values [][]byte
	for key, value := range datastore {
		userKeys = append(userKeys, key)
		values = append(values, value)
	}

	userlib.DatastoreSet(userKeys[0], values[1])
	i := 1
	for i < len(userKeys) {
		userlib.DatastoreSet(userKeys[i], values[0])
		i += 1
	}

	err = u1.AppendFile("file1", appendContents)
	if err == nil {
		t.Error("Should not be able to append")
		return
	}

	userlib.DatastoreClear()
	userlib.KeystoreClear()

	datastore = userlib.DatastoreGetMap()

	files := []string{"file1", "file2", "file3", "file4", "file5"}
	users := []string{"user1", "user2", "user3", "user4", "user5"}
	for i, offset := range []int{-15, -2, 0, 4, 9} {
		user, err := InitUser(users[i], "pass")
		if err != nil {
			t.Error(err)
			return
		}

		fileContents := userlib.RandomBytes(userlib.AESBlockSize * 3 - offset)
		appendContents := userlib.RandomBytes(userlib.AESBlockSize)
		user.StoreFile(files[i], fileContents)
		err = user.AppendFile(files[i], appendContents)
		if err != nil {
			t.Error("Couldn't append on normal procedure")
			return
		}

		contentsGet, err := user.LoadFile(files[i])
		if err != nil {
			t.Error(err)
			return
		}
		finalOut := append(fileContents, appendContents...)
		if !reflect.DeepEqual(finalOut, contentsGet) {
			t.Error("Files don't match", finalOut, contentsGet)
			return
		}

		var userKeys []userlib.UUID
		var values [][]byte
		for key, value := range datastore {
			userKeys = append(userKeys, key)
			values = append(values, value)
		}

		hasError := false
		for key := range userKeys {
			datastore[userKeys[key]] = userlib.RandomBytes(len(values[key]))
			contentsGet, err = user.LoadFile(files[i])
			if err != nil {
				hasError = true
			}
			datastore[userKeys[key]] = values[key]
		}

		if !hasError {
			t.Error("Corrupted datastore but no failed file load.")
		}
	}
}

func TestShare(t *testing.T) {
	userlib.SetDebugStatus(true)
	clear()
	u, err := InitUser("alice", "fubar")
	if err != nil {
		t.Error("Failed to initialize user", err)
		return
	}
	u2, err2 := InitUser("bob", "foobar")
	if err2 != nil {
		t.Error("Failed to initialize bob", err2)
		return
	}

	v := []byte("This is a test")
	u.StoreFile("file1", v)

	var v2 []byte
	var magic_string string

	v, err = u.LoadFile("file1")
	if err != nil {
		t.Error("Failed to download the file from alice", err)
		return
	}

	magic_string, err = u.ShareFile("file1", "bob")
	if err != nil {
		t.Error("Failed to share the a file", err)
		return
	}
	err = u2.ReceiveFile("file2", "alice", magic_string)
	if err != nil {
		t.Error("Failed to receive the share message", err)
		return
	}

	v2, err = u2.LoadFile("file2")
	if err != nil {
		t.Error("Failed to download the file after sharing", err)
		return
	}
	if !reflect.DeepEqual(v, v2) {
		t.Error("Shared file is not the same", v, v2)
		return
	}

	v2, err = u.LoadFile("file1")
	if err != nil {
		t.Error("Failed to download the file after sharing", err)
		return
	}
	if !reflect.DeepEqual(v, v2) {
		t.Error("Shared file is not the same", v, v2)
		return
	}

	// Need to test storing a file with the same name as a file that has been shared with that person.
	v = []byte("This should also affect u1")
	u2.StoreFile("file2", v)
	v2, err = u2.LoadFile("file2")
	if err != nil {
		t.Error("Failed to download the file after sharing", err)
		return
	}
	if !reflect.DeepEqual(v, v2) {
		t.Error("Shared file is not the same", v, v2)
		return
	}

	v2, err = u.LoadFile("file1")
	if err != nil {
		t.Error("Failed to download the file after sharing", err)
		return
	}
	if !reflect.DeepEqual(v, v2) {
		t.Error("Shared file is not the same", v, v2)
		return
	}

	// Append share
	v = []byte("This should also affect u1 but append")
	vAppended := []byte("This should also affect u1This should also affect u1 but append")
	u.AppendFile("file1", v)
	v2, err = u2.LoadFile("file2")
	if err != nil {
		t.Error("Failed to download the file after sharing", err)
		return
	}
	if !reflect.DeepEqual(vAppended, v2) {
		t.Error("Shared file is not the same", v, v2)
		return
	}

	v2, err = u.LoadFile("file1")
	if err != nil {
		t.Error("Failed to download the file after sharing", err)
		return
	}
	if !reflect.DeepEqual(vAppended, v2) {
		t.Error("Shared file is not the same", v, v2)
		return
	}

}

func TestShareBadMagicString(t *testing.T) {
	userlib.SetDebugStatus(true)
	userlib.DatastoreClear()
	userlib.KeystoreClear()
	datastore := userlib.DatastoreGetMap()
	keystore := userlib.KeystoreGetMap()
	_, _ = datastore, keystore

	u1, err := InitUser("user1", "pass")
	if err != nil {
		t.Error("Could not create user1", err)
		return
	}
	u2, err := InitUser("user2", "pass")
	if err != nil {
		t.Error("Could not create user2", err)
		return
	}

	fileContents := userlib.RandomBytes(userlib.AESBlockSize)
	u1.StoreFile("file1", fileContents)
	_, err = u1.LoadFile("file1")
	if err != nil {
		t.Error("Could not load file1", err)
		return
	}

	magic_string, err := u1.ShareFile("file1", "user2")
	if err != nil {
		t.Error("Could not share file1 with user2", err)
		return
	}

	i := 1
	for i < len(magic_string) {
		bad_magic_string := magic_string[:i]
		err = u2.ReceiveFile("file2", "user1", bad_magic_string)
		if err == nil {
			t.Error("Should not have been able to receive bad_magic_string")
			return
		}
		insertGarbage := userlib.RandomBytes(i)
		bad_magic_string = magic_string[:i] + string(insertGarbage) + magic_string[i:]
		err = u2.ReceiveFile("file2", "user1", bad_magic_string)
		if err == nil {
			t.Error("Should not have been able to receive bad_magic_string")
			return
		}
		bad_magic_string = magic_string[:(i/2)] + magic_string[(len(magic_string)-(i/2)):]
		err = u2.ReceiveFile("file2", "user1", bad_magic_string)
		if err == nil {
			t.Error("Should not have been able to receive bad_magic_string")
			return
		}
		i += 5
	}

	// magic string is empty. Should error
	bad_magic_string := ""
	err = u2.ReceiveFile("file2", "user1", bad_magic_string)
	if err == nil {
		t.Error("Magic string that is empty should error")
		return
	}

	// Some garbage magic string that should error
	len_string := len(magic_string)
	bad_magic_string = string(userlib.RandomBytes(len_string))
	err = u2.ReceiveFile("file2", "user1", bad_magic_string)
	if err == nil {
		t.Error("Garbage magic string should error")
		return
	}
}

func TestShareNoAccessUser(t *testing.T) {
	userlib.SetDebugStatus(true)
	userlib.DatastoreClear()
	userlib.KeystoreClear()
	datastore := userlib.DatastoreGetMap()
	keystore := userlib.KeystoreGetMap()
	_, _ = datastore, keystore

	u1, err := InitUser("user1", "pass")
	if err != nil {
		t.Error("Failed to create user1", err)
		return
	}

	u2, err := InitUser("user2", "pass")
	if err != nil {
		t.Error("Failed to create user2", err)
		return
	}

	u3, err := InitUser("user3", "pass")
	if err != nil {
		t.Error("Failed to create user3", err)
		return
	}

	fileContents := userlib.RandomBytes(userlib.AESBlockSize)
	u1.StoreFile("file1", fileContents)

	magic_string, err := u1.ShareFile("file1", "user2")
	if err != nil {
		t.Error("Failed to share file1", err)
		return
	}


	// U2 has access
	err = u2.ReceiveFile("file2", "user1", magic_string)
	if err != nil {
		t.Error("Could not receive message", err)
		return
	}

	// U3 should not have access
	err = u3.ReceiveFile("file3", "user1", magic_string)
	if err == nil {
		t.Error("Should not have been able to receive message", err)
		return
	}

	_, err = u3.LoadFile("file3")
	if err == nil {
		t.Error("User3 does not have access but still loads", err)
		return
	}

}

func TestShareNoOwnership(t *testing.T) {
	userlib.SetDebugStatus(true)
	userlib.DatastoreClear()
	userlib.KeystoreClear()
	datastore := userlib.DatastoreGetMap()
	keystore := userlib.KeystoreGetMap()
	_, _ = datastore, keystore

	u1, err := InitUser("user1", "pass")
	if err != nil {
		t.Error("Failed to create user1", err)
		return
	}

	_, err = u1.ShareFile("file1", "user2")
	if err == nil {
		t.Error("Failed to error. U1 doesn't have file1")
	}
}

func TestMultipleShare(t *testing.T) {
	userlib.SetDebugStatus(false)
	clear()
	u, err := InitUser("alice", "fubar")
	if err != nil {
		t.Error("Failed to initialize user", err)
		return
	}
	u2, err2 := InitUser("bob", "foobar")
	if err2 != nil {
		t.Error("Failed to initialize bob", err2)
		return
	}
	u3, err3 := InitUser("charlie", "fomo")
	if err3 != nil {
		t.Error("Failed to initialize charlie", err2)
		return
	}
	u4, err4 := InitUser("delta", "fodo")
	if err4 != nil {
		t.Error("Failed to initialize delta", err2)
		return
	}

	v := []byte("This is a test")
	u.StoreFile("file1", v)

	var v2 []byte
	var magic_string string

	v, err = u.LoadFile("file1")
	if err != nil {
		t.Error("Failed to download the file from alice", err)
		return
	}

	magic_string, err = u.ShareFile("file1", "bob")
	if err != nil {
		t.Error("Failed to share the a file", err)
		return
	}
	err = u2.ReceiveFile("file2", "alice", magic_string)
	if err != nil {
		t.Error("Failed to receive the share message", err)
		return
	}
	magic_string, err = u.ShareFile("file1", "charlie")
	if err != nil {
		t.Error("Failed to share the a file", err)
		return
	}
	err = u3.ReceiveFile("file3", "alice", magic_string)
	if err != nil {
		t.Error("Failed to receive the share message", err)
		return
	}
	magic_string, err = u.ShareFile("file1", "delta")
	if err != nil {
		t.Error("Failed to share the a file2", err)
		return
	}
	err = u4.ReceiveFile("file4", "alice", magic_string)
	if err != nil {
		t.Error("Failed to receive the share message", err)
		return
	}

	v2, err = u2.LoadFile("file2")
	if err != nil {
		t.Error("Failed to download the file after sharing", err)
		return
	}
	if !reflect.DeepEqual(v, v2) {
		t.Error("Shared file is not the same", v, v2)
		return
	}

	v2, err = u.LoadFile("file1")
	if err != nil {
		t.Error("Failed to download the file after sharing", err)
		return
	}
	if !reflect.DeepEqual(v, v2) {
		t.Error("Shared file is not the same", v, v2)
		return
	}
	v2, err = u3.LoadFile("file3")
	if err != nil {
		t.Error("Failed to download the file after sharing", err)
		return
	}
	if !reflect.DeepEqual(v, v2) {
		t.Error("Shared file is not the same", v, v2)
		return
	}
	v2, err = u4.LoadFile("file4")
	if err != nil {
		t.Error("Failed to download the file after sharing", err)
		return
	}
	if !reflect.DeepEqual(v, v2) {
		t.Error("Shared file is not the same", v, v2)
		return
	}
}

func TestReceiveOverwriteExisting(t *testing.T) {
	userlib.SetDebugStatus(true)
	userlib.DatastoreClear()
	userlib.KeystoreClear()
	datastore := userlib.DatastoreGetMap()
	keystore := userlib.KeystoreGetMap()
	_, _ = datastore, keystore

	u1, err := InitUser("user1", "pass")
	if err != nil {
		t.Error("Failed to create user1", err)
		return
	}

	u2, err := InitUser("user2", "pass")
	if err != nil {
		t.Error("Failed to create user2", err)
		return
	}

	fileContents1 := userlib.RandomBytes(userlib.AESBlockSize)
	u1.StoreFile("file1", fileContents1)

	fileContents2 := userlib.RandomBytes(userlib.AESBlockSize)
	u2.StoreFile("file2", fileContents2)

	magic_string, err := u1.ShareFile("file1", "user2")
	if err != nil {
		t.Error("Failed to share file1", err)
		return
	}

	err = u2.ReceiveFile("file2", "user1", magic_string)
	if err == nil {
		t.Error("Should not be able to receive message", err)
		return
	}
}

func TestRevoke(t *testing.T) {
	userlib.SetDebugStatus(true)
	clear()
	u, err := InitUser("alice", "fubar")
	if err != nil {
		t.Error("Failed to initialize user", err)
		return
	}
	u2, err2 := InitUser("bob", "foobar")
	if err2 != nil {
		t.Error("Failed to initialize bob", err2)
		return
	}

	v := []byte("This is a test")
	u.StoreFile("file1", v)

	var v2 []byte
	var magic_string string

	v, err = u.LoadFile("file1")
	if err != nil {
		t.Error("Failed to download the file from alice", err)
		return
	}

	magic_string, err = u.ShareFile("file1", "bob")
	if err != nil {
		t.Error("Failed to share the a file", err)
		return
	}
	err = u2.ReceiveFile("file2", "alice", magic_string)
	if err != nil {
		t.Error("Failed to receive the share message", err)
		return
	}

	v2, err = u2.LoadFile("file2")
	if err != nil {
		t.Error("Failed to download the file after sharing", err)
		return
	}
	if !reflect.DeepEqual(v, v2) {
		t.Error("Shared file is not the same", v, v2)
		return
	}

	v2, err = u.LoadFile("file1")
	if err != nil {
		t.Error("Failed to download the file after sharing", err)
		return
	}
	if !reflect.DeepEqual(v, v2) {
		t.Error("Shared file is not the same", v, v2)
		return
	}

	err = u.RevokeFile("file1", "bob")
	if err != nil {
		t.Error("Revoke failed", err)
	}

	v1 := []byte("Append")
	vAppended := []byte("This is a testAppend")
	u.AppendFile("file1", v1)
	v2, err = u.LoadFile("file1")
	if err != nil {
		t.Error("Failed to download the file after sharing", err)
		return
	}
	if !reflect.DeepEqual(vAppended, v2) {
		t.Error("Shared file is not the same", v, v2)
		return
	}

	v2, err = u2.LoadFile("file2")
	if err == nil {
		t.Error("Bob can access old file")
	}

}

func TestMaliciousRevoke(t *testing.T) {
	clear()
	u, err := InitUser("alice", "fubar")
	if err != nil {
		t.Error("Failed to initialize user", err)
		return
	}
	u2, err2 := InitUser("bob", "foobar")
	if err2 != nil {
		t.Error("Failed to initialize bob", err2)
		return
	}

	v := []byte("This is a test")
	u.StoreFile("file1", v)

	var v2 []byte
	var magic_string string

	v, err = u.LoadFile("file1")
	if err != nil {
		t.Error("Failed to download the file from alice", err)
		return
	}

	magic_string, err = u.ShareFile("file1", "bob")
	if err != nil {
		t.Error("Failed to share the a file", err)
		return
	}
	err = u2.ReceiveFile("file2", "alice", magic_string)
	if err != nil {
		t.Error("Failed to receive the share message", err)
		return
	}
	v2, err = u2.LoadFile("file2")
	if err != nil {
		t.Error("Failed to download the file after sharing", err)
		return
	}
	if !reflect.DeepEqual(v, v2) {
		t.Error("Shared file is not the same", v, v2)
		return
	}

	v2, err = u.LoadFile("file1")
	if err != nil {
		t.Error("Failed to download the file after sharing", err)
		return
	}
	if !reflect.DeepEqual(v, v2) {
		t.Error("Shared file is not the same", v, v2)
		return
	}
	err = u.RevokeFile("file2", "alice")
	if err == nil {
		t.Error("Revoke success", err)
		return
	}
}

func TestBadRevoke(t *testing.T) {
	userlib.SetDebugStatus(true)
	clear()
	u, err := InitUser("one", "foobar")
	if err != nil {
		t.Error("Failed to initialize user", err)
		return
	}
	u2, err2 := InitUser("two", "foobar")
	if err2 != nil {
		t.Error("Failed to initialize bob", err2)
		return
	}
	u3, err3 := InitUser("three", "foobar")
	if err3 != nil {
		t.Error("Failed to initialize charlie", err2)
		return
	}

	v := []byte("This is a test")
	u.StoreFile("file1", v)

	magic_string, err := u.ShareFile("file1", "two")
	if err != nil {
		t.Error("Failed to share the a file", err)
		return
	}
	err = u2.ReceiveFile("file2", "one", magic_string)
	if err != nil {
		t.Error("Failed to receive the share message", err)
		return
	}

	magic_string, err = u2.ShareFile("file2", "three")
	if err != nil {
		t.Error("Failed to share the a file", err)
		return
	}
	err = u3.ReceiveFile("file3", "two", magic_string)
	if err != nil {
		t.Error("Failed to receive the share message", err)
		return
	}

	err = u.RevokeFile("file1", "three")
	if err == nil {
		t.Error("Revoke did not error", err)
	}

}

func TestTreeRevoke(t *testing.T) {
	userlib.SetDebugStatus(false)
	clear()
	u, err := InitUser("alice", "fubar")
	if err != nil {
		t.Error("Failed to initialize user", err)
		return
	}
	u2, err2 := InitUser("bob", "foobar")
	if err2 != nil {
		t.Error("Failed to initialize bob", err2)
		return
	}
	u3, err3 := InitUser("charlie", "fomo")
	if err3 != nil {
		t.Error("Failed to initialize charlie", err2)
		return
	}
	u4, err4 := InitUser("delta", "fodo")
	if err4 != nil {
		t.Error("Failed to initialize delta", err2)
		return
	}

	v := []byte("This is a test")
	u.StoreFile("file1", v)

	var v2 []byte
	var magic_string string

	v, err = u.LoadFile("file1")
	if err != nil {
		t.Error("Failed to download the file from alice", err)
		return
	}

	magic_string, err = u.ShareFile("file1", "bob")
	if err != nil {
		t.Error("Failed to share the a file", err)
		return
	}
	err = u2.ReceiveFile("file2", "alice", magic_string)
	if err != nil {
		t.Error("Failed to receive the share message", err)
		return
	}
	magic_string, err = u.ShareFile("file1", "charlie")
	if err != nil {
		t.Error("Failed to share the a file", err)
		return
	}
	err = u3.ReceiveFile("file3", "alice", magic_string)
	if err != nil {
		t.Error("Failed to receive the share message", err)
		return
	}
	magic_string, err = u2.ShareFile("file2", "delta")
	if err != nil {
		t.Error("Failed to share the a file2", err)
		return
	}
	err = u4.ReceiveFile("file4", "bob", magic_string)
	if err != nil {
		t.Error("Failed to receive the share message", err)
		return
	}

	v2, err = u2.LoadFile("file2")
	if err != nil {
		t.Error("Failed to download the file after sharing", err)
		return
	}
	if !reflect.DeepEqual(v, v2) {
		t.Error("Shared file is not the same", v, v2)
		return
	}

	v2, err = u.LoadFile("file1")
	if err != nil {
		t.Error("Failed to download the file after sharing", err)
		return
	}
	if !reflect.DeepEqual(v, v2) {
		t.Error("Shared file is not the same", v, v2)
		return
	}
	v2, err = u3.LoadFile("file3")
	if err != nil {
		t.Error("Failed to download the file after sharing", err)
		return
	}
	if !reflect.DeepEqual(v, v2) {
		t.Error("Shared file is not the same", v, v2)
		return
	}
	v2, err = u4.LoadFile("file4")
	if err != nil {
		t.Error("Failed to download the file after sharing", err)
		return
	}
	if !reflect.DeepEqual(v, v2) {
		t.Error("Shared file is not the same", v, v2)
		return
	}

	err = u.RevokeFile("file1", "bob")
	if err != nil {
		t.Error("Revoke failed", err)
	}

	v2, err = u.LoadFile("file1")
	if err != nil {
		t.Error("Failed to download the file from alice after revoke", err)
	}
	if !reflect.DeepEqual(v, v2) {
		t.Error("Revoked file is not the same", v, v2)
	}

	_, err = u2.LoadFile("file2")
	if err == nil {
		t.Error("Bob can still access the revoked file")
	}
	_, err = u4.LoadFile("file4")
	if err == nil {
		t.Error("Delta can still access the revoked file")
	}

	fileContentAppend := []byte("End of File")
	u2.AppendFile("file2", fileContentAppend)
	if err == nil {
		t.Error("Bob can still append")
	}
	u4.AppendFile("file4", fileContentAppend)
	if err == nil {
		t.Error("Delta can still append")
	}


	v2, err = u3.LoadFile("file3")
	if err != nil {
		t.Error("Failed to download the file after sharing", err)
		return
	}
	if !reflect.DeepEqual(v, v2) {
		t.Error("Shared file is not the same", v, v2)
		return
	}
}

func TestNamedShare(t *testing.T) {
	userlib.SetDebugStatus(true)
	clear()
	u, err := InitUser("alice", "fubar")
	if err != nil {
		t.Error("Failed to initialize user", err)
		return
	}
	u2, err2 := InitUser("bob", "foobar")
	if err2 != nil {
		t.Error("Failed to initialize bob", err2)
		return
	}

	v := []byte("This is a test")
	v2 := []byte("This is taken")
	u.StoreFile("file1", v)
	u2.StoreFile("file2", v2)

	var magic_string string

	v, err = u.LoadFile("file1")
	if err != nil {
		t.Error("Failed to download the file from alice", err)
		return
	}

	v2, err = u2.LoadFile("file2")
	if err != nil {
		t.Error("Failed to download the file from bob", err)
		return
	}

	magic_string, err = u.ShareFile("file1", "bob")
	if err != nil {
		t.Error("Failed to share the a file", err)
		return
	}
	err = u2.ReceiveFile("file2", "alice", magic_string)
	if err == nil {
		t.Error("Failed to check filename", err)
		return
	}
}
