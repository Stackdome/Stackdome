package services

import (
	"encoding/base64"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("AESEncryptionService", func() {
	var (
		service EncryptionService
		err     error
	)

	const masterKey = "this-is-a-very-secure-master-key-that-is-at-least-64-characters-long-for-security-validation"

	BeforeEach(func() {
		service, err = NewAESEncryptionService(EncryptionServiceSpec{Masterkey: masterKey})
		Expect(err).ToNot(HaveOccurred())
		Expect(service).ToNot(BeNil())
	})

	Context("when encrypting and decrypting data", func() {
		It("should return the original plaintext after encryption and decryption", func() {
			plaintext := []byte("Hello, this is a secret message!")
			encrypted, err := service.EncryptData(plaintext)
			Expect(err).ToNot(HaveOccurred())
			Expect(encrypted).ToNot(BeEmpty())

			decrypted, err := service.DecryptData(encrypted)
			Expect(err).ToNot(HaveOccurred())
			Expect(decrypted).To(Equal(plaintext))
		})
	})

	Context("when encrypting invalid input", func() {
		It("should return an error when input is empty", func() {
			encrypted, err := service.EncryptData([]byte{})
			Expect(err).To(HaveOccurred())
			Expect(encrypted).To(BeEmpty())
		})
	})

	Context("when decrypting invalid input", func() {
		It("should return an error for empty input", func() {
			_, err := service.DecryptData("")
			Expect(err).To(HaveOccurred())
		})

		It("should return an error for base64 invalid input", func() {
			_, err := service.DecryptData("not-base64!!!")
			Expect(err).To(HaveOccurred())
		})

		It("should return an error for truncated encrypted data", func() {
			badEncrypted := base64.StdEncoding.EncodeToString([]byte("short"))
			_, err := service.DecryptData(badEncrypted)
			Expect(err).To(HaveOccurred())
		})

		It("should fail to decrypt tampered ciphertext", func() {
			plaintext := []byte("Top secret")
			encrypted, err := service.EncryptData(plaintext)
			Expect(err).ToNot(HaveOccurred())

			decoded, err := base64.StdEncoding.DecodeString(encrypted)
			Expect(err).ToNot(HaveOccurred())
			decoded[len(decoded)-1] ^= 0xFF // Flip a bit

			tampered := base64.StdEncoding.EncodeToString(decoded)
			_, err = service.DecryptData(tampered)
			Expect(err).To(HaveOccurred())
		})
	})

	Context("when creating service with invalid master key", func() {
		It("should return an error for short master key", func() {
			service, err := NewAESEncryptionService(EncryptionServiceSpec{Masterkey: "short-key"})
			Expect(err).To(HaveOccurred())
			Expect(service).To(BeNil())
		})
	})

	Context("multiple encryption", func() {
		It("should produce different ciphertexts for same input", func() {
			plaintext := []byte("repeatable input")
			encrypted1, err1 := service.EncryptData(plaintext)
			encrypted2, err2 := service.EncryptData(plaintext)

			Expect(err1).ToNot(HaveOccurred())
			Expect(err2).ToNot(HaveOccurred())

			Expect(encrypted1).ToNot(Equal(encrypted2)) // Different salt/nonce = different ciphertext
		})
	})
})
