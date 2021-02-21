package proj2

// CS 161 Project 2 Fall 2020
// You MUST NOT change what you import.  If you add ANY additional
// imports it will break the autograder. We will be very upset.

import (
	// You neet to add with
	// go get github.com/cs161-staff/userlib
	"github.com/cs161-staff/userlib"

	// Life is much easier with json:  You are
	// going to want to use this so you can easily
	// turn complex structures into strings etc...
	"encoding/json"

	// Likewise useful for debugging, etc...
	"encoding/hex"

	// UUIDs are generated right based on the cryptographic PRNG
	// so lets make life easier and use those too...
	//
	// You need to add with "go get github.com/google/uuid"
	"github.com/google/uuid"

	// Useful for debug messages, or string manipulation for datastore keys.
	"strings"

	// Want to import errors.
	"errors"

	// Optional. You can remove the "_" there, but please do not touch
	// anything else within the import bracket.
	_ "strconv"

	// if you are looking for fmt, we don't give you fmt, but you can use userlib.DebugMsg.
	// see someUsefulThings() below:
)

// This serves two purposes:
// a) It shows you some useful primitives, and
// b) it suppresses warnings for items not being imported.
// Of course, this function can be deleted.
func someUsefulThings() {
	// Creates a random UUID
	f := uuid.New()
	userlib.DebugMsg("UUID as string:%v", f.String())

	// Example of writing over a byte of f
	f[0] = 10
	userlib.DebugMsg("UUID as string:%v", f.String())

	// takes a sequence of bytes and renders as hex
	h := hex.EncodeToString([]byte("fubar"))
	userlib.DebugMsg("The hex: %v", h)

	// Marshals data into a JSON representation
	// Will actually work with go structures as well
	d, _ := json.Marshal(f)
	userlib.DebugMsg("The json data: %v", string(d))
	var g uuid.UUID
	json.Unmarshal(d, &g)
	userlib.DebugMsg("Unmashaled data %v", g.String())

	// This creates an error type
	userlib.DebugMsg("Creation of error %v", errors.New(strings.ToTitle("This is an error")))

	// And a random RSA key.  In this case, ignoring the error
	// return value
	var pk userlib.PKEEncKey
        var sk userlib.PKEDecKey
	pk, sk, _ = userlib.PKEKeyGen()
	userlib.DebugMsg("Key is %v, %v", pk, sk)
}

// Helper function: Takes the first 16 bytes and
// converts it into the UUID type
func bytesToUUID(data []byte) (ret uuid.UUID) {
	for x := range ret {
		ret[x] = data[x]
	}
	return
}

// The structure definition for a user record
type User struct {
	Username string
	UUID userlib.UUID
	UPC []byte
	SymEncKey []byte
	HMacKey []byte

	PrivateRSAKey userlib.PKEDecKey
	PrivateSignKey userlib.DSSignKey

	OwnedFiles map[string]userlib.UUID
	AccessibleFiles map[string]userlib.UUID

	// You can add other fields here if you want...
	// Note for JSON to marshal/unmarshal, the fields need to
	// be public (start with a capital letter)
}

type Container struct {
	CipherData []byte
	HMAC []byte
}

type FileInfo struct {
	HLoc userlib.UUID
	FKey []byte
	FHMACKey []byte
	Children map[string]userlib.UUID

}

type SharedFileInfo struct {
	ILoc userlib.UUID
	SKey []byte
	SHMACKey []byte

}

type ChildInfo struct {
	ILoc userlib.UUID
	SKey []byte
	SHMACKey []byte

}

type FileHeader struct {
	filename string
	FilePieces []userlib.UUID

}

type FilePiece struct {
	FileData []byte

}

func GenerateUserInformation(username string, password string, userdata *User) {
	usernameBytes := []byte(username)
	passwordBytes := []byte(password)
	userdata.UPC = userlib.Argon2Key(passwordBytes, usernameBytes, 32)
	HMAC_UPC, _ := userlib.HMACEval(userdata.UPC, usernameBytes)
	userdata.UUID, _ = uuid.FromBytes(HMAC_UPC[:16])
	userdata.SymEncKey = userlib.Argon2Key(append([]byte("encrypt_"), userdata.UPC...),
		passwordBytes, uint32(userlib.AESKeySize))
	userdata.HMacKey = userlib.Argon2Key(append([]byte("mac_"), userdata.UPC...),
		passwordBytes, uint32(userlib.AESKeySize))
	return
}

func StoreInDataStore(uuid uuid.UUID, hMacKey []byte, symEncKey []byte, dataToStore []byte) (err error){
	iv := userlib.RandomBytes(userlib.AESBlockSize)
	cipherData := SymEncrypt(symEncKey, iv, dataToStore)
	containerCipherData, err := Containerize(hMacKey, cipherData)
	if err != nil { return }
	containerCipherDataBytes, err := json.Marshal(containerCipherData)
	userlib.DatastoreSet(uuid, containerCipherDataBytes)
	return
}

func GetFromDataStore(uuid uuid.UUID, hmacKey []byte, symEncKey []byte) (decryptedData []byte, err error) {
	containerizedCipherDataBytes, success := userlib.DatastoreGet(uuid)
	if !success {
		err = errors.New("No UUID found")
		return
	}

	var container Container
	err = json.Unmarshal(containerizedCipherDataBytes, &container)
	if err != nil {
		return
	}
	cipherData, err := DeContainerize(hmacKey, container)
	if err != nil { return }

	decryptedData = SymDecrypt(symEncKey, cipherData)
	return
}

func Containerize(key []byte, cipherData []byte) (container Container, err error) {
	container = Container{cipherData, make([]byte, len(cipherData))}
	container.HMAC, err = userlib.HMACEval(key, cipherData)
	return
}

func DeContainerize(key []byte, container Container) (cipherData []byte, err error) {
	HMAC_Result, err := userlib.HMACEval(key, container.CipherData)
	if !userlib.HMACEqual(container.HMAC, HMAC_Result) {
		err = errors.New("Cannot decontainerize")
		return
	}
	cipherData = container.CipherData
	return
}

func SymEncrypt(key []byte, iv []byte, dataToEncrpyt []byte) (cipherData []byte) {
		paddedDataToEncrypt := pad(dataToEncrpyt)
		encrypedData := []byte(userlib.SymEnc(key, iv, paddedDataToEncrypt))
		cipherData = encrypedData
		return
}

func SymDecrypt(key []byte, cipherData []byte) (data []byte) {

		cipherCombinedDecrypted := userlib.SymDec(key, cipherData)
		cipherCombinedDecrypted = unPad(cipherCombinedDecrypted)
		data = cipherCombinedDecrypted
		return
}

func (userdata *User) StoreUser() (err error) {
	userDataBytes, _ := json.Marshal(userdata)
	err = StoreInDataStore(userdata.UUID, userdata.HMacKey, userdata.SymEncKey, userDataBytes)
	return
}

func (userdata *User) UpdateUser() (user *User) {
	bytesUserdata, err := GetFromDataStore(userdata.UUID, userdata.HMacKey, userdata.SymEncKey)
	if err != nil { return userdata }

	err = json.Unmarshal(bytesUserdata, userdata)
	if err != nil { return userdata }
	return userdata
}

func pad(text []byte) (paddedText []byte) {
	padCount := userlib.AESBlockSize - (len(text) % userlib.AESBlockSize)
	paddedText = make([]byte, padCount+len(text))
	for i := 0; i <= len(text); i++ {
		if i == len(text) {
			paddedText[i] = 1
		} else {
			paddedText[i] = (text)[i]
		}
	}
	return
}

func unPad(text []byte) (unpaddedText []byte) {
	var padBegin uint
	for padBegin = uint(len(text) - 1); padBegin >= 0 && text[padBegin] == 0; padBegin -- {
	}
	unpaddedText = text[:padBegin]
	return
}

func RandUUID() (UUID userlib.UUID) {
	for true {
		UUID = uuid.New()
		_, ok := userlib.DatastoreGet(UUID)
		if !ok {
			break
		}
	}
	return UUID
}

// This creates a user.  It will only be called once for a user
// (unless the keystore and datastore are cleared during testing purposes)

// It should store a copy of the userdata, suitably encrypted, in the
// datastore and should store the user's public key in the keystore.

// The datastore may corrupt or completely erase the stored
// information, but nobody outside should be able to get at the stored

// You are not allowed to use any global storage other than the
// keystore and the datastore functions in the userlib library.

// You can assume the password has strong entropy, EXCEPT
// the attackers may possess a precomputed tables containing
// hashes of common passwords downloaded from the internet.
func InitUser(username string, password string) (userdataptr *User, err error) {
	if username == "" {
		return nil, errors.New("Username can't be empty")
	}
	if password == "" {
		return nil, errors.New("Password can't be empty")
	}
	var userdata User
	userdataptr = &userdata
	GenerateUserInformation(username, password, &userdata)

	userdata.Username = username
	userdata.OwnedFiles = make(map[string]userlib.UUID)
	userdata.AccessibleFiles = make(map[string]userlib.UUID)

	ek, dk, _ := userlib.PKEKeyGen()
	userlib.KeystoreSet("encrypt_" + username, ek)
	userdata.PrivateRSAKey = dk

	sk, vk, _ := userlib.DSKeyGen()
	userlib.KeystoreSet("verify_" + username, vk)
	userdata.PrivateSignKey = sk

	err = userdata.StoreUser()
	if err != nil { return nil, err}

	return
}

// This fetches the user information from the Datastore.  It should
// fail with an error if the user/password is invalid, or if the user
// data was corrupted, or if the user can't be found.
func GetUser(username string, password string) (userdataptr *User, err error) {
	var userdata User
	userdataptr = &userdata
	GenerateUserInformation(username, password, &userdata)

	dataBytes, err := GetFromDataStore(userdataptr.UUID, userdataptr.HMacKey, userdataptr.SymEncKey)
	if err != nil {
		if err.Error() == "No UUID found" {
			err = errors.New("Username or Password is incorrect")
		}
		return
	}
	err = json.Unmarshal(dataBytes, userdataptr)
	if err != nil { return nil,err}
	return
}

// This stores a file in the datastore.
//
// The plaintext of the filename + the plaintext and length of the filename
// should NOT be revealed to the datastore!
func (userdata *User) StoreFile(filename string, data []byte) {
	var (
		FPiece FilePiece
		FHeader FileHeader
		FInfo FileInfo
	)

	userdata = userdata.UpdateUser()

	if _, ok := userdata.OwnedFiles[filename]; ok {
		//Create FilePiece
		//Put data in FilePiece
		FPiece.FileData = data

		//Get FileInfo and decrypt
		FInfoUUID := userdata.OwnedFiles[filename]
		if len(userdata.HMacKey) == 0 || len(userdata.SymEncKey) == 0 {
			return
		}
		bytesFInfo, err := GetFromDataStore(FInfoUUID, userdata.HMacKey, userdata.SymEncKey)
		err = json.Unmarshal(bytesFInfo,&FInfo)
		if err != nil { return }

		//Get FileHeader and decrypt
		if len(FInfo.FKey) == 0 || len(FInfo.FHMACKey) == 0 {
			return
		}
		bytesFHeader, err := GetFromDataStore(FInfo.HLoc, FInfo.FHMACKey, FInfo.FKey)
		err = json.Unmarshal(bytesFHeader,&FHeader)
		if err != nil { return }

		//Get FilePiece and decrypt
		//For each FilePiece in FilePieces, decrypt
		for i := range FHeader.FilePieces {
			userlib.DatastoreDelete(FHeader.FilePieces[i])
		}

		FHeader.FilePieces = make([]uuid.UUID, 1)

		//Encrypt FilePiece with FKey and FMACKey (randomly made) and store
		bytesFPiece, err := json.Marshal(FPiece)
		if err != nil { return }
		FPieceUUID := RandUUID()
		err = StoreInDataStore(FPieceUUID, FInfo.FHMACKey, FInfo.FKey, bytesFPiece)
		if err != nil { return }

		//Put location of FilePiece in the FilePieces in FileHeader
		FHeader.FilePieces[0] = FPieceUUID

		//Encrypt FileHeader with FKey and FMACKey (randomly made) and store
		bytesFHeader, err = json.Marshal(FHeader)
		if err != nil { return }
		// FHeaderUUID := RandUUID()
		err = StoreInDataStore(FInfo.HLoc, FInfo.FHMACKey, FInfo.FKey, bytesFHeader)
		if err != nil { return }
		err = userdata.StoreUser()
		if err != nil { return }
		return
	}
	if _, ok := userdata.AccessibleFiles[filename]; ok {
		var IFS FileInfo
		//Create FilePiece
		//Put data in FilePiece
		FPiece.FileData = data

		//Get FileInfo and decrypt
		FInfoUUID := userdata.AccessibleFiles[filename]
		if len(userdata.HMacKey) == 0 || len(userdata.SymEncKey) == 0 {
			return
		}
		bytesFInfo, err := GetFromDataStore(FInfoUUID, userdata.HMacKey, userdata.SymEncKey)
		if err != nil { return }
		err = json.Unmarshal(bytesFInfo,&FInfo)
		if err != nil { return }

		//Get InfoForSharing and decrypt
		if len(FInfo.FHMACKey) == 0 || len(FInfo.FKey) == 0 {
			return
		}
		bytesIFS, err := GetFromDataStore(FInfo.HLoc, FInfo.FHMACKey, FInfo.FKey)
		if err != nil { return }
		err = json.Unmarshal(bytesIFS,&IFS)
		if err != nil { return }

		//Get FileHeader and decrypt
		if len(IFS.FHMACKey) == 0 || len(IFS.FKey) == 0 {
			return
		}
		bytesFHeader, err := GetFromDataStore(IFS.HLoc, IFS.FHMACKey, IFS.FKey)
		if err != nil { return }
		err = json.Unmarshal(bytesFHeader,&FHeader)
		if err != nil { return }

		//Get FilePiece and decrypt
		//For each FilePiece in FilePieces, decrypt
		for i := range FHeader.FilePieces {
			userlib.DatastoreDelete(FHeader.FilePieces[i])
		}

		FHeader.FilePieces = make([]uuid.UUID, 1)

		//Encrypt FilePiece with FKey and FMACKey (randomly made) and store
		bytesFPiece, err := json.Marshal(FPiece)
		if err != nil { return }
		FPieceUUID := RandUUID()
		err = StoreInDataStore(FPieceUUID, IFS.FHMACKey, IFS.FKey, bytesFPiece)
		if err != nil { return }

		//Put location of FilePiece in the FilePieces in FileHeader
		FHeader.FilePieces[0] = FPieceUUID

		//Encrypt FileHeader with FKey and FMACKey (randomly made) and store
		bytesFHeader, err = json.Marshal(FHeader)
		if err != nil { return }
		err = StoreInDataStore(IFS.HLoc, IFS.FHMACKey, IFS.FKey, bytesFHeader)
		if err != nil { return }
		err = userdata.StoreUser()
		return
	}
	if _, ok := userdata.OwnedFiles[filename]; !ok {
		//Create FilePiece
		//Put data in FilePiece
		FPiece.FileData = data
		//Create FileHeader
		FHeader.FilePieces = make([]uuid.UUID, 1)
		//Create FileInfo
		//Create FKey and FMACKey and store in FileInfo
		FKey := userlib.RandomBytes(32)
		FHMACKey := userlib.RandomBytes(32)
		FInfo.FKey = FKey
		FInfo.FHMACKey = FHMACKey

		FInfo.Children = make(map[string]userlib.UUID)

		//Encrypt FilePiece with FKey and FMACKey (randomly made) and store
		bytesFPiece, err := json.Marshal(FPiece)
		if err != nil { return }
		FPieceUUID := RandUUID()
		if len(FInfo.FKey) == 0 || len(FInfo.FHMACKey) == 0 {
			return
		}
		err = StoreInDataStore(FPieceUUID, FInfo.FHMACKey, FInfo.FKey, bytesFPiece)
		if err != nil { return }

		//Put location of FilePiece in the FilePieces in FileHeader
		FHeader.FilePieces[0] = FPieceUUID

		//Encrypt FileHeader with FKey and FMACKey (randomly made) and store
		bytesFHeader, err := json.Marshal(FHeader)
		if err != nil { return }
		FHeaderUUID := RandUUID()
		err = StoreInDataStore(FHeaderUUID, FInfo.FHMACKey, FInfo.FKey, bytesFHeader)
		if err != nil { return }
		//Put location of FileHeader in HLoc in FileInfo
		FInfo.HLoc = FHeaderUUID

		//Encrypt FileInfo with original user related stuff
		bytesFInfo, err := json.Marshal(FInfo)
		if err != nil { return }
		FInfoUUID := RandUUID()
		if len(userdata.HMacKey) == 0 || len(userdata.SymEncKey) == 0 {
			return
		}
		err = StoreInDataStore(FInfoUUID, userdata.HMacKey, userdata.SymEncKey, bytesFInfo)
		if err != nil { return }

		//Add FileInfo to OwnedFiles with Hash
		userdata.OwnedFiles[filename] = FInfoUUID

		//Restore userdata
		err = userdata.StoreUser()
		if err != nil { return }
		// TESTING PURPOSE
		// data, _ := GetFromDataStore(FPieceUUID, FInfo.FHMACKey, FInfo.FKey)
		// var FPieceNew FilePiece
		// _ = json.Unmarshal(data, &FPieceNew)
		// userlib.DebugMsg(string(FPieceNew.FileData))
	}

	return
}

// This adds on to an existing file.
//
// Append should be efficient, you shouldn't rewrite or reencrypt the
// existing file, but only whatever additional information and
// metadata you need.
func (userdata *User) AppendFile(filename string, data []byte) (err error) {
	var (
		FInfo FileInfo
		FHeader FileHeader
		FPiece FilePiece
		FInfoUUID uuid.UUID
	)
	userdata = userdata.UpdateUser()

	//If in Owned Files
	if _, ok := userdata.OwnedFiles[filename]; ok {
		FInfoUUID = userdata.OwnedFiles[filename]
		//Setup new FilePiece
		FPiece.FileData = data

		//Get FileInfo and decrypt
		if len(userdata.HMacKey) == 0 || len(userdata.SymEncKey) == 0 {
			return errors.New("No HMACkey or FKey")
		}
		bytesFInfo, err := GetFromDataStore(FInfoUUID, userdata.HMacKey, userdata.SymEncKey)
		if err != nil { return err }
		err = json.Unmarshal(bytesFInfo, &FInfo)
		if err != nil { return err }

		//Encrypt FilePiece with FKey and FMACKey (randomly made) and store
		bytesFPiece, err := json.Marshal(FPiece)
		if err != nil { return err }
		FPieceUUID := RandUUID()
		if len(FInfo.FHMACKey) == 0 || len(FInfo.FKey) == 0 {
			return errors.New("No HMACkey or FKey")
		}
		err = StoreInDataStore(FPieceUUID, FInfo.FHMACKey, FInfo.FKey, bytesFPiece)
		if err != nil { return err }

		//Get FileHeader and decrypt
		bytesFHeader, err := GetFromDataStore(FInfo.HLoc, FInfo.FHMACKey, FInfo.FKey)
		if err != nil { return err }
		err = json.Unmarshal(bytesFHeader, &FHeader)
		if err != nil { return err }

		//Add a new FilePiece into FilePieces
		FHeader.FilePieces = append(FHeader.FilePieces, FPieceUUID)

		//Restore FileHeader
		bytesFHeader, err = json.Marshal(FHeader)
		if err != nil { return err }
		err = StoreInDataStore(FInfo.HLoc, FInfo.FHMACKey, FInfo.FKey, bytesFHeader)
		if err != nil { return err }
	//If in Accessible FIles
	} else if _, ok := userdata.AccessibleFiles[filename] ; ok {
		var IFS FileInfo
		FInfoUUID = userdata.AccessibleFiles[filename]
		//Setup new FilePiece
		FPiece.FileData = data

		//Get FileInfo and decrypt
		if len(userdata.HMacKey) == 0 || len(userdata.SymEncKey) == 0 {
			return errors.New("No HMACkey or FKey")
		}
		bytesFInfo, err := GetFromDataStore(FInfoUUID, userdata.HMacKey, userdata.SymEncKey)
		if err != nil { return err }
		err = json.Unmarshal(bytesFInfo, &FInfo)
		if err != nil { return err }

		//Get InfoForSharing and decrypt
		if len(FInfo.FHMACKey) == 0 || len(FInfo.FKey) == 0 {
			return errors.New("No HMACkey or FKey")
		}
		bytesIFS, err := GetFromDataStore(FInfo.HLoc, FInfo.FHMACKey, FInfo.FKey)
		err = json.Unmarshal(bytesIFS,&IFS)
		if err != nil { return err }

		//Encrypt FilePiece with FKey and FMACKey (randomly made) and store
		bytesFPiece, err := json.Marshal(FPiece)
		if err != nil { return err }
		FPieceUUID := RandUUID()
		if len(IFS.FHMACKey) == 0 || len(IFS.FKey) == 0 {
			return errors.New("No HMACkey or FKey")
		}
		err = StoreInDataStore(FPieceUUID, IFS.FHMACKey, IFS.FKey, bytesFPiece)
		if err != nil { return err }

		//Get FileHeader and decrypt
		bytesFHeader, err := GetFromDataStore(IFS.HLoc, IFS.FHMACKey, IFS.FKey)
		if err != nil { return err }
		err = json.Unmarshal(bytesFHeader, &FHeader)
		if err != nil { return err }

		//Add a new FilePiece into FilePieces
		FHeader.FilePieces = append(FHeader.FilePieces, FPieceUUID)

		//Restore FileHeader
		bytesFHeader, err = json.Marshal(FHeader)
		if err != nil { return err }
		err = StoreInDataStore(IFS.HLoc, IFS.FHMACKey, IFS.FKey, bytesFHeader)
		if err != nil { return err }
	} else {
		err = errors.New("Cannot find file")
		return err
	}
	err = userdata.StoreUser()
	if err != nil { return err }

	return
}

// This loads a file from the Datastore.
//
// It should give an error if the file is corrupted in any way.
func (userdata *User) LoadFile(filename string) (data []byte, err error) {
	var (
		FInfo FileInfo
		FHeader FileHeader
		FPiece FilePiece
	)
	userdata = userdata.UpdateUser()

	//Check Owned Files
	if _, ok := userdata.OwnedFiles[filename] ; ok {
		//Get FileInfo and decrypt
		FInfoUUID := userdata.OwnedFiles[filename]
		if len(userdata.HMacKey) == 0 || len(userdata.SymEncKey) == 0 {
			return nil, errors.New("No key")
		}
		bytesFInfo, err := GetFromDataStore(FInfoUUID, userdata.HMacKey, userdata.SymEncKey)
		if err != nil {
				return nil, errors.New("Cannot load file")
			}
		err = json.Unmarshal(bytesFInfo,&FInfo)
		//Get FileHeader and decrypt
		if len(FInfo.FHMACKey) == 0 || len(FInfo.FKey) == 0 {
			return nil, errors.New("No key")
		}
		bytesFHeader, err := GetFromDataStore(FInfo.HLoc, FInfo.FHMACKey, FInfo.FKey)
		if err != nil {
				return nil, errors.New("Cannot load file")
			}
		err = json.Unmarshal(bytesFHeader,&FHeader)
		if err != nil {
				return nil, errors.New("Cannot unmarshal")
			}
		//Get FilePiece and decrypt
		//For each FilePiece in FilePieces, decrypt
		var bytesFPieces []byte
		for i := range FHeader.FilePieces {
			var data []byte
			data, err = GetFromDataStore(FHeader.FilePieces[i], FInfo.FHMACKey, FInfo.FKey)
			if err != nil {
				return nil, errors.New("Cannot load file")
			}
			err = json.Unmarshal(data, &FPiece)
			if err != nil {
				return nil, errors.New("Cannot unmarshal")
			}
			bytesFPieces = append(bytesFPieces, FPiece.FileData...)
		}
		return bytesFPieces, nil
	//Check Accessible Files
	} else if _, ok := userdata.AccessibleFiles[filename] ; ok {
		var IFS FileInfo
		//Get SFileInfo and decrypt
		FInfoUUID := userdata.AccessibleFiles[filename]
		if len(userdata.HMacKey) == 0 || len(userdata.SymEncKey) == 0 {
			return nil, errors.New("No key")
		}
		bytesFInfo, err := GetFromDataStore(FInfoUUID, userdata.HMacKey, userdata.SymEncKey)
		if err != nil {
			return nil, errors.New("Cannot load file")
		}
		err = json.Unmarshal(bytesFInfo,&FInfo)
		if err != nil {
			return nil, errors.New("Cannot unmarshal")
		}
		//Get SInfoForSharing and decrypt
		if len(FInfo.FHMACKey) == 0 || len(FInfo.FKey) == 0 {
			return nil, errors.New("No key")
		}
		bytesIFS, err := GetFromDataStore(FInfo.HLoc, FInfo.FHMACKey, FInfo.FKey)
		if err != nil {
			return nil, errors.New("Cannot load file")
		}
		err = json.Unmarshal(bytesIFS,&IFS)
		if err != nil {
			return nil, errors.New("Cannot unmarshal")
		}
		//Get FileHeader and decrypt
		if len(IFS.FHMACKey) == 0 || len(IFS.FKey) == 0 {
			return nil, errors.New("No key")
		}
		bytesFHeader, err := GetFromDataStore(IFS.HLoc, IFS.FHMACKey, IFS.FKey)
		if err != nil {
			return nil, errors.New("Cannot load file")
		}
		err = json.Unmarshal(bytesFHeader,&FHeader)
		if err != nil {
			return nil, errors.New("Cannot unmarshal")
		}
		//Get FilePiece and decrypt
		//For each FilePiece in FilePieces, decrypt
		var bytesFPieces []byte
		for i := range FHeader.FilePieces {
			var data []byte
			data, err = GetFromDataStore(FHeader.FilePieces[i], IFS.FHMACKey, IFS.FKey)
			if err != nil {
				return nil, errors.New("Cannot load file")
			}
			err = json.Unmarshal(data, &FPiece)
			if err != nil {
				return nil, errors.New("Cannot unmarshal")
			}
			bytesFPieces = append(bytesFPieces, FPiece.FileData...)
		}
		return bytesFPieces, nil
	//Else, error
	} else {
		return nil, errors.New("No file")
	}
	return
}

type ContainerShare struct {
	CipherData []byte
	HMAC []byte
	SymKey []byte
}
// This creates a sharing record, which is a key pointing to something
// in the datastore to share with the recipient.

// This enables the recipient to access the encrypted file as well
// for reading/appending.

// Note that neither the recipient NOR the datastore should gain any
// information about what the sender calls the file.  Only the
// recipient can access the sharing record, and only the recipient
// should be able to know the sender.
func (userdata *User) ShareFile(filename string, recipient string) (
	magic_string string, err error) {
	var (
		FInfo FileInfo


		InfoForSharing FileInfo
		ChildInfo FileInfo
	)
	userdata = userdata.UpdateUser()

	// If you own the file
	if _, ok := userdata.OwnedFiles[filename] ; ok {

		//Get FileInfo and decrypt
		FInfoUUID := userdata.OwnedFiles[filename]
		if len(userdata.HMacKey) == 0 || len(userdata.SymEncKey) == 0 {
			return "", errors.New("No key")
		}
		bytesFInfo, err := GetFromDataStore(FInfoUUID, userdata.HMacKey, userdata.SymEncKey)
		err = json.Unmarshal(bytesFInfo,&FInfo)
		if err != nil { return "", err }

		//Create InfoForSharing and encrypt and then store
		InfoForSharing.HLoc = FInfo.HLoc
		InfoForSharing.FKey = FInfo.FKey
		InfoForSharing.FHMACKey = FInfo.FHMACKey
		InfoForSharing.Children = make(map[string]userlib.UUID)
		IFSUUID := RandUUID()

		//Create ChildInfo and encrypt and then store
		ChildInfo.HLoc = IFSUUID
		FKey := userlib.RandomBytes(32)
		FHMACKey := userlib.RandomBytes(32)
		ChildInfo.FKey = FKey
		ChildInfo.FHMACKey = FHMACKey
		ChildInfo.Children = make(map[string]userlib.UUID)
		bytesCInfo, err := json.Marshal(ChildInfo)
		if err != nil { return "", err }
		CInfoUUID := RandUUID()
		err = StoreInDataStore(CInfoUUID, userdata.HMacKey, userdata.SymEncKey, bytesCInfo)
		if err != nil { return "", err }

		//Reference ChildInfo in FileInfo
		FInfo.Children[recipient] = CInfoUUID
		bytesInfoNew, err := json.Marshal(FInfo)
		//Re encrypt FileInfo
		err = StoreInDataStore(FInfoUUID, userdata.HMacKey, userdata.SymEncKey, bytesInfoNew)
		if err != nil { return "", err }

		//Encrypt InfoForSharing
		bytesIFS, err := json.Marshal(InfoForSharing)
		if err != nil { return "", err }
		err = StoreInDataStore(IFSUUID, ChildInfo.FHMACKey, ChildInfo.FKey, bytesIFS)
		if err != nil { return "", err }

		//Send the magic string by encrypting and signing
		recKey, ok := userlib.KeystoreGet("encrypt_" + recipient)
		if !ok { return "", err }

		symEncKey := userlib.RandomBytes(userlib.AESBlockSize)
		iv := userlib.RandomBytes(userlib.AESBlockSize)
		cipherData := SymEncrypt(symEncKey, iv, bytesCInfo)

		// var encryptedSend []byte
		//
		// i := 0
		// for i < len(bytesCInfo) {
		// 	var data []byte
		// 	if i + 64 >= len(bytesCInfo) {
		// 		data, err = userlib.PKEEnc(recKey, bytesCInfo[i:])
		// 		if err != nil { return "", err }
		// 		encryptedSend = append(encryptedSend, data...)
		// 		break
		// 	}
		// 	userlib.DebugMsg("EnterShare")
		// 	data, err = userlib.PKEEnc(recKey, bytesCInfo[i:i+64])
		// 	if err != nil { return "", err }
		// 	encryptedSend = append(encryptedSend, data...)
		// 	i += 64
		// 	userlib.DebugMsg("ExitShare")
		// }

		symEncKeyPlusPublicEnc, _ := userlib.PKEEnc(recKey, symEncKey)
		verifiedSend, err := userlib.DSSign(userdata.PrivateSignKey, cipherData)
		if err != nil { return "", err }

		//Container it
		container := ContainerShare{cipherData, make([]byte, len(cipherData)), symEncKeyPlusPublicEnc}
		container.HMAC = verifiedSend
		bytesContainer, err := json.Marshal(container)
		if err != nil { return "", err }
		err = userdata.StoreUser()
		if err != nil { return "", err }
		return string(bytesContainer), nil
	} else if _, ok := userdata.AccessibleFiles[filename] ; ok {

		//Get FileInfo and decrypt
		FInfoUUID := userdata.AccessibleFiles[filename]
		if len(userdata.HMacKey) == 0 || len(userdata.SymEncKey) == 0 {
			return "", errors.New("No key")
		}
		bytesFInfo, err := GetFromDataStore(FInfoUUID, userdata.HMacKey, userdata.SymEncKey)
		if err != nil { return "", err }
		err = json.Unmarshal(bytesFInfo,&FInfo)
		if err != nil { return "", err }

		//Send the magic string by encrypting and signing
		recKey, ok := userlib.KeystoreGet("encrypt_" + recipient)
		if !ok { return "", err }

		symEncKey := userlib.RandomBytes(userlib.AESBlockSize)
		iv := userlib.RandomBytes(userlib.AESBlockSize)
		cipherData := SymEncrypt(symEncKey, iv, bytesFInfo)
		//encryptedSend, err := userlib.PKEEnc(recKey, bytesFInfo)
		//if err != nil { return "", err }
		//verifiedSend, err := userlib.DSSign(userdata.PrivateSignKey, encryptedSend)
		//if err != nil { return "", err }

		symEncKeyPlusPublicEnc, _ := userlib.PKEEnc(recKey, symEncKey)
		verifiedSend, err := userlib.DSSign(userdata.PrivateSignKey, cipherData)
		if err != nil { return "", err }

		//Container it
		container := ContainerShare{cipherData, make([]byte, len(cipherData)), symEncKeyPlusPublicEnc}
		container.HMAC = verifiedSend
		bytesContainer, err := json.Marshal(container)
		if err != nil { return "", err }
		return string(bytesContainer), err
	}
	//If you don't have access to the file
	return "", errors.New("file doesn't exist")
}

// Note recipient's filename can be different from the sender's filename.
// The recipient should not be able to discover the sender's view on
// what the filename even is!  However, the recipient must ensure that
// it is authentically from the sender.
func (userdata *User) ReceiveFile(filename string, sender string,
	magic_string string) error {
	var (
		SInfo FileInfo
		container ContainerShare
	)
	userdata = userdata.UpdateUser()

	if _, ok := userdata.OwnedFiles[filename]; ok {
		return errors.New("filename exists")
	}
	if _, ok := userdata.AccessibleFiles[filename]; ok {
		return errors.New("filename exists")
	}

	//Uncontainer
	err := json.Unmarshal([]byte(magic_string), &container)
	if err != nil { return err }

	if err != nil { return err }
	//Decrypt magic_string into SInfo
	sendVerify, ok := userlib.KeystoreGet("verify_" + sender)
	if !ok { return err }

	verify := userlib.DSVerify(sendVerify, container.CipherData, container.HMAC)
	if verify != nil { return verify }
	symEncKeyRec, _ := userlib.PKEDec(userdata.PrivateRSAKey, container.SymKey)

	if len(symEncKeyRec) == 0 {
		return errors.New("Could not retrieve decrypt key")
	}
	bytesInfo := SymDecrypt(symEncKeyRec, container.CipherData)

	// bytesInfo, err := userlib.PKEDec(userdata.PrivateRSAKey, container.CipherData)
	err = json.Unmarshal(bytesInfo, &SInfo)
	if err != nil { return err }

	//Encrypt SharedFileInfo with recipient user related stuff
	SInfoUUID := RandUUID()
	err = StoreInDataStore(SInfoUUID, userdata.HMacKey, userdata.SymEncKey, bytesInfo)
	if err != nil { return err }
	//Add FileInfo to OwnedFiles with Hash
	userdata.AccessibleFiles[filename] = SInfoUUID

	//Restore userdata
	err = userdata.StoreUser()
	if err != nil { return err }

	return err
}

// Removes target user's access.
func (userdata *User) RevokeFile(filename string, target_username string) (err error) {
	var (
		FInfo FileInfo
		FHeader FileHeader
	)
	userdata = userdata.UpdateUser()
	//Somewhere, re encrypt FilePiece, FileHeader, FileInfo

	// If you own the file
	if _, ok := userdata.OwnedFiles[filename] ; ok {

		NewFKey := userlib.RandomBytes(32)
		NewFHMACKey := userlib.RandomBytes(32)

		//Get FileInfo and decrypt
		FInfoUUID := userdata.OwnedFiles[filename]
		if len(userdata.HMacKey) == 0 || len(userdata.SymEncKey) == 0 {
			return errors.New("No key")
		}
		bytesFInfo, err := GetFromDataStore(FInfoUUID, userdata.HMacKey, userdata.SymEncKey)
		if err != nil {return err}
		err = json.Unmarshal(bytesFInfo,&FInfo)
		if err != nil {return err}

		//Get FileHeader and decrypt
		if len(FInfo.FHMACKey) == 0 || len(FInfo.FKey) == 0 {
			return errors.New("No key")
		}
		bytesFHeader, err := GetFromDataStore(FInfo.HLoc, FInfo.FHMACKey, FInfo.FKey)
		if err != nil {return err}
		err = json.Unmarshal(bytesFHeader,&FHeader)
		if err != nil {return err}

		//Get FilePiece and decrypt
		//For each FilePiece in FilePieces, re encrypt
		for i := range FHeader.FilePieces {
			var data []byte
			data, err = GetFromDataStore(FHeader.FilePieces[i], FInfo.FHMACKey, FInfo.FKey)
			if err != nil {
				return errors.New("Cannot load file")
			}
			err = StoreInDataStore(FHeader.FilePieces[i], NewFHMACKey, NewFKey, data)
			if err != nil {
				return errors.New("Cannot store file")
			}
		}

		//Re encrypt FileHeader
		newbytesFHeader, err := json.Marshal(FHeader)
		if err != nil {return err}
		err = StoreInDataStore(FInfo.HLoc, NewFHMACKey, NewFKey, newbytesFHeader)
		if err != nil {return err}

		target_exist := false
		//For each child, change the InfoForSharing
		for key, _ := range FInfo.Children {
			var CInfo FileInfo
			var IFS FileInfo
			childData, err := GetFromDataStore(FInfo.Children[key], userdata.HMacKey, userdata.SymEncKey)
			if err != nil {
				return err
			}
			err = json.Unmarshal(childData, &CInfo)
			if err != nil {return err}

			//If the child is the one you want to revoke
			if key == target_username {
				target_exist = true
				userlib.DatastoreDelete(CInfo.HLoc)
				userlib.DatastoreDelete(FInfo.Children[key])
				delete(FInfo.Children, key)

			//Else get the InfoForSharing
			} else {
				if len(CInfo.FHMACKey) == 0 || len(CInfo.FKey) == 0 {
					return errors.New("No key")
				}
				IFSData, err := GetFromDataStore(CInfo.HLoc, CInfo.FHMACKey, CInfo.FKey)
				if err != nil {
					return errors.New("Cannot load file")
				}
				err = json.Unmarshal(IFSData, &IFS)
				if err != nil {return err}
				IFS.FKey = NewFKey
				IFS.FHMACKey = NewFHMACKey
				IFS.Children =  make(map[string]userlib.UUID)
				bytesIFS, err := json.Marshal(IFS)
				if err != nil {return err}
				err = StoreInDataStore(CInfo.HLoc, CInfo.FHMACKey, CInfo.FKey, bytesIFS)
				if err != nil {return err}
			}
		}

		if target_exist == false {
			return errors.New("target user never existed")
		}

		FInfo.FKey = NewFKey
		FInfo.FHMACKey = NewFHMACKey

		//Re encrypt the FileInfo
		newbytesFInfo, err := json.Marshal(FInfo)
		if err != nil {return err}
		err = StoreInDataStore(FInfoUUID, userdata.HMacKey, userdata.SymEncKey, newbytesFInfo)
		if err != nil {return err}

		return err
	}
	return errors.New("No owned file")
}
