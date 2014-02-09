package coin

import (
    "bytes"
    "errors"
    "github.com/skycoin/skycoin/src/lib/base58"
    "log"
)

var (
    addressVersions = map[string]byte{
        "main": 0x0F, //main network
        "test": 0x1F, //test network
    }
    // Address version is a global default version used for all address
    // creation and checking
    addressVersion = addressVersions["test"]
)

// Returns the named address version and whether it is a known version
func VersionByName(name string) (byte, bool) {
    v, ok := addressVersions[name]
    return v, ok
}

// Returns the named address version, panics if unknown name
func MustVersionByName(name string) byte {
    v, ok := VersionByName(name)
    if !ok {
        log.Panicf("Invalid version name: %s", name)
    }
    return v
}

// Sets the address version used for all address creation and checking
func SetAddressVersion(name string) {
    addressVersion = MustVersionByName(name)
    logger.Debug("Set address version to %s", name)
}

type Checksum [4]byte

//version is after Key to enable better vanity address generation
//Address stuct is a 25 byte with a 20 byte publickey hash, 1 byte address
//type and 4 byte checksum.
type Address struct {
    Key     [20]byte //20 byte pubkey hash
    Version byte     //1 byte
    ChkSum  [4]byte  //4 byte checksum, first 4 bytes of sha256 of key+version
}

// Creates Address from PubKey as ripemd160(sha256(sha256(pubkey)))
func AddressFromPubKey(pubKey PubKey) Address {
    addr := Address{
        Version: addressVersion,
        Key:     pubKey.ToAddressHash(),
    }
    addr.setChecksum()
    return addr
}

// Creates an Address from its base58 encoding.  Will panic if the addr is
// invalid
func MustDecodeBase58Address(addr string) Address {
    a, err := addressFromBytes(base58.Base582Hex(addr))
    if err != nil {
        log.Panicf("Invalid address %s", addr)
    }
    return a
}

// Creates an Address from its base58 encoding
func DecodeBase58Address(addr string) (Address, error) {
    return addressFromBytes(base58.Base582Hex(addr))
}

// Returns an address given an Address.Bytes()
func addressFromBytes(b []byte) (Address, error) {
    var a Address
    keyLen := len(a.Key)
    if len(b) != keyLen+len(a.ChkSum)+1 {
        return a, errors.New("Invalid address bytes")
    }
    copy(a.Key[:], b[:keyLen])
    a.Version = b[keyLen]
    copy(a.ChkSum[:], b[keyLen+1:])
    if !a.IsValidChecksum() {
        return a, errors.New("Invalid checksum")
    } else {
        return a, nil
    }
}

// Checks that the address appears valid for the public key
func (self *Address) Verify(key PubKey) error {
    if self.Key != key.ToAddressHash() {
        return errors.New("Public key invalid for address")
    }
    if self.Version != addressVersion {
        return errors.New("Invalid address version")
    }
    if !self.IsValidChecksum() {
        return errors.New("Invalid address checksum")
    }
    return nil
}

// Address as Base58 encoded string
func (self *Address) String() string {
    return string(base58.Hex2Base58(self.Bytes()))
}

// Returns address as raw bytes, containing version and then key
func (self *Address) Bytes() []byte {
    keyLen := len(self.Key)
    b := make([]byte, keyLen+len(self.ChkSum)+1)
    copy(b[:keyLen], self.Key[:])
    b[keyLen] = self.Version
    copy(b[keyLen+1:], self.ChkSum[:])
    return b
}

// Returns Address Checksum which is the first 4 bytes of sha256(key+version)
func (self *Address) Checksum() Checksum {
    // Version comes after the address to support vanity addresses
    r1 := append(self.Key[:], []byte{self.Version}...)
    r2 := SumSHA256(r1[:])
    var c Checksum
    copy(c[:], r2[:len(c)])
    return c
}

// Returns whether the checksum on address is valid for its key
func (self *Address) IsValidChecksum() bool {
    c := self.Checksum()
    return bytes.Equal(c[:], self.ChkSum[:])
}

func (self *Address) setChecksum() {
    self.ChkSum = self.Checksum()
}
