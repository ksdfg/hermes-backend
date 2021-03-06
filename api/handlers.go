package api

import (
	"bytes"
	"encoding/base64"
	"encoding/csv"
	"errors"
	"fmt"
	"image/png"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"hermes/config"

	"github.com/Rhymen/go-whatsapp"
	"github.com/boombuler/barcode"
	"github.com/boombuler/barcode/qr"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type message struct {
	recipient string
	body      string
	sent      bool
	log       string
}

type sessionData struct {
	conn       *whatsapp.Conn
	Processing bool     `json:"processing"`
	Success    []string `json:"success,omitempty"`
	Failures   []string `json:"failures,omitempty"`
}

type service struct {
	config  *config.Config
	session map[uuid.UUID]sessionData
}

// new will start a fresh session
func (svc service) new(ctx *fiber.Ctx) error {
	// Create new whatsapp connection
	conn, err := whatsapp.NewConn(20 * time.Second)
	if err != nil {
		return err
	}

	// Set client version and name
	conn.SetClientVersion(svc.config.VersionMajor, svc.config.VersionMinor, svc.config.VersionPatch)
	err = conn.SetClientName(svc.config.ClientLong, svc.config.ClientShort, svc.config.ClientVersion)
	if err != nil {
		return err
	}

	// Create new session ID
	id, err := uuid.NewUUID()
	if err != nil {
		log.Println(err)
		return err
	}

	// Set connection and list of messages in service session map
	svc.session[id] = sessionData{conn: conn}

	// Set session ID in context headers
	ctx.Set("session", id.String())

	// Get auth code for login
	qrChan := make(chan string)
	go func() {
		_, err = conn.Login(qrChan)
		if err != nil {
			log.Println(err)

			// Disconnect the whatsapp connection, not required anymore
			_, err = conn.Disconnect()
			if err != nil {
				log.Println(err)
			}
		}
	}()
	authCode, _ := <-qrChan
	close(qrChan)

	// Encode auth code into a QR for user to scan
	qrCode, err := qr.Encode(authCode, qr.L, qr.Auto)
	if err != nil {
		return err
	}
	qrCode, err = barcode.Scale(qrCode, svc.config.QrSize, svc.config.QrSize)
	if err != nil {
		return err
	}

	// Encode QR into a base64 png image
	var buf bytes.Buffer
	err = png.Encode(&buf, qrCode)
	qrString := base64.StdEncoding.EncodeToString(buf.Bytes())

	return ctx.SendString(fmt.Sprintf("data:image/png;base64,%s", qrString))
}

// loggedIn will return true if the user in the session is logged in, else false
func (svc service) loggedIn(ctx *fiber.Ctx) error {
	// Get session ID from request context
	idStr := ctx.Get("session")
	if idStr == "" {
		return errors.New("session ID not passed")
	}
	id, err := uuid.Parse(idStr)
	if err != nil {
		log.Println(err)
		return err
	}

	// Get data from service session map
	data, ok := svc.session[id]
	if !ok {
		return errors.New("data not set in session")
	}

	return ctx.JSON(map[string]bool{"loggedIn": data.conn.GetLoggedIn()})
}

// send accepts a message body template and a CSV with the phone number and fields to replace
func (svc service) send(ctx *fiber.Ctx) error {
	// Get session ID from request context
	idStr := ctx.Get("session")
	if idStr == "" {
		return errors.New("session ID not passed")
	}
	id, err := uuid.Parse(idStr)
	if err != nil {
		log.Println(err)
		return err
	}

	// Get data from service session map
	data, ok := svc.session[id]
	if !ok {
		return errors.New("data not set in session")
	}

	// Get body to be sent in message
	body := strings.TrimSpace(ctx.FormValue("body"))
	if body == "" {
		log.Println("body not specified in request form")
		return errors.New("body not specified in request form")
	}

	// Fetch uploaded csv file
	header, err := ctx.FormFile("file")
	if err != nil {
		log.Println(err)
		return err
	}
	file, err := header.Open()
	if err != nil {
		log.Println(err)
		return err
	}

	// Read rows from CSV
	reader := csv.NewReader(file)
	rows, err := reader.ReadAll()
	if err != nil {
		log.Println(err)
		return err
	}

	// Get indexes of required headers
	phoneIndex := -1
	headers := rows[0]
	for i, header := range headers {
		if header == "phone" {
			phoneIndex = i
			break
		}
	}
	if phoneIndex == -1 {
		log.Println("Header 'phone' not found in CSV")
		return errors.New("header 'phone' not found in CSV")
	}

	// Get messages from the rows of the CSV
	var messages []message
	for _, row := range rows[1:] {
		// Parse message body
		messageBody := body
		for i, header := range headers {
			messageBody = strings.ReplaceAll(messageBody, fmt.Sprintf("{{%s}}", strings.TrimSpace(header)), strings.TrimSpace(row[i]))
		}

		// Append message
		messages = append(messages, message{
			recipient: strings.TrimSpace(row[phoneIndex]),
			body:      messageBody,
		})
	}

	// Goroutine to get logs from Send operations and set them in session data
	logChan := make(chan message)
	go func() {
		for m := range logChan {
			if m.sent {
				data.Success = append(data.Success, m.log)
			} else {
				data.Failures = append(data.Failures, m.log)
			}
			svc.session[id] = data
		}
	}()

	// Send messages in goroutine
	go func() {
		// Set processing flag as true
		data.Processing = true
		svc.session[id] = data

		// Declare wait group to handle execution of batched goroutines
		var wg sync.WaitGroup

		// Iterate over all messages in batches
		for i := 0; i < len(messages); i += svc.config.Concurrency {
			// Calculate end index for splice
			j := i + svc.config.Concurrency
			if j > len(messages) {
				j = len(messages)
			}

			// Set delta to number of goroutines spawned
			wg.Add(j - i)

			// Iterate over all messages in the splice and send them in goroutines
			for _, msg := range messages[i:j] {
				go func(msg message) {
					// Send message
					_, err = data.conn.Send(whatsapp.TextMessage{
						Info: whatsapp.MessageInfo{RemoteJid: fmt.Sprintf("%s@s.whatsapp.net", msg.recipient)},
						Text: msg.body,
					})
					if err != nil {
						msg.log = fmt.Sprintf("Could not send message to %s: %s", msg.recipient, err)
					} else {
						msg.log = fmt.Sprintf("Message sent to %s", msg.recipient)
						msg.sent = true
					}

					// Send logs to channel
					logChan <- msg

					// Reduce delta by 1
					wg.Done()
				}(msg)
			}

			// Wait till all goroutines end execution
			wg.Wait()
		}

		// Close logs channel after all messages are sent
		close(logChan)

		// Logout user after messages are sent
		err = data.conn.Logout()
		if err != nil {
			log.Println(err)
		}

		// Disconnect the whatsapp connection, not required anymore
		_, err = data.conn.Disconnect()
		if err != nil {
			log.Println(err)
		}

		// Set processing flag as false
		data.Processing = false
		svc.session[id] = data
	}()

	return ctx.SendStatus(http.StatusOK)
}

// logs will return all the logs for the send operation triggered for the session
func (svc service) logs(ctx *fiber.Ctx) error {
	// Get session ID from request context
	idStr := ctx.Get("session")
	if idStr == "" {
		return errors.New("session ID not passed")
	}
	id, err := uuid.Parse(idStr)
	if err != nil {
		log.Println(err)
		return err
	}

	// Get data from service session map
	data, ok := svc.session[id]
	if !ok {
		return errors.New("data not set in session")
	}

	return ctx.JSON(data)
}

func (svc service) cleanup(ctx *fiber.Ctx) error {
	// Get session ID from request context
	idStr := ctx.Get("session")
	if idStr == "" {
		return errors.New("session ID not passed")
	}
	id, err := uuid.Parse(idStr)
	if err != nil {
		log.Println(err)
		return err
	}

	// Delete session data
	delete(svc.session, id)

	return ctx.SendStatus(http.StatusOK)
}
