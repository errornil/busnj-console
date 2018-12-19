package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"sort"
	"strings"
)

type ReceiptFile struct {
	Uuid            string
	RequestDateMs   string
	HasBinary       bool
	HasDebugJson    bool
	ReceiptResponse ReceiptResponse
}

type byDate []ReceiptFile

func (s byDate) Len() int {
	return len(s)
}
func (s byDate) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}
func (s byDate) Less(i, j int) bool {
	return s[i].RequestDateMs > s[j].RequestDateMs
	// todo: compare int, not string
}

type byTransactionId []InApp

func (s byTransactionId) Len() int {
	return len(s)
}
func (s byTransactionId) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}
func (s byTransactionId) Less(i, j int) bool {
	return s[i].TransactionId > s[j].TransactionId
	// todo: compare int, not string
}

type TemplateDataTransactions struct {
	InApp map[string]InApp
}

type TemplateDataHistory struct {
	Receipts []ReceiptFile
}

type ReceiptResponse struct {
	Status             int                  `json:"status"`
	Environment        string               `json:"environment"`
	Receipt            Receipt              `json:"receipt"`
	PendingRenewalInfo []PendingRenewalInfo `json:"pending_renewal_info"`
}

type PendingRenewalInfo struct {
	ExpirationIntent       string `json:"expiration_intent"`
	AutoRenewProductId     string `json:"auto_renew_product_id"`
	OriginalTransactionId  string `json:"original_transaction_id"`
	IsInBillingRetryPeriod string `json:"is_in_billing_retry_period"`
	ProductId              string `json:"product_id"`
	AutoRenewStatus        string `json:"auto_renew_status"`
}

type Receipt struct {
	ReceiptType                string      `json:"receipt_type"`
	AdamId                     int         `json:"adam_id"`
	AppItemId                  interface{} `json:"app_item_id"`
	BundleId                   string      `json:"bundle_id"`
	ApplicationVersion         string      `json:"application_version"`
	DownloadId                 int         `json:"download_id"`
	VersionExternalIdentifier  interface{} `json:"version_external_identifier"`
	ReceiptCreationDate        string      `json:"receipt_creation_date"`
	ReceiptCreationDateMs      string      `json:"receipt_creation_date_ms"`
	ReceiptCreationDatePst     string      `json:"receipt_creation_date_pst"`
	RequestDate                string      `json:"request_date"`
	RequestDateMs              string      `json:"request_date_ms"`
	RequestDatePst             string      `json:"request_date_pst"`
	OriginalPurchaseDate       string      `json:"original_purchase_date"`
	OriginalPurchaseDate_ms    string      `json:"original_purchase_date_ms"`
	OriginalPurchaseDate_pst   string      `json:"original_purchase_date_pst"`
	OriginalApplicationVersion string      `json:"original_application_version"`
	InApp                      []InApp     `json:"in_app"`
}

type InApp struct {
	Quantity                string `json:"quantity"`
	ProductId               string `json:"product_id"`
	TransactionId           string `json:"transaction_id"`
	OriginalTransactionId   string `json:"original_transaction_id"`
	PurchaseDate            string `json:"purchase_date"`
	PurchaseDateMs          string `json:"purchase_date_ms"`
	PurchaseDatePst         string `json:"purchase_date_pst"`
	OriginalPurchaseDate    string `json:"original_purchase_date"`
	OriginalPurchaseDateMs  string `json:"original_purchase_date_ms"`
	OriginalPurchaseDatePst string `json:"original_purchase_date_pst"`
	ExpiresDate             string `json:"expires_date"`
	ExpiresDateMs           string `json:"expires_date_ms"`
	ExpiresDatePst          string `json:"expires_date_pst"`
	WebOrderLineItemId      string `json:"web_order_line_item_id"`
	IsTrialPeriod           string `json:"is_trial_period"`
	IsInIntroOfferPeriod    string `json:"is_in_intro_offer_period"`

	ReceiptUuids []string
	ReceiptFile  ReceiptFile
}

func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
		panic(fmt.Sprintf("%s: %s", msg, err))
	}
}

func handlerHistory(w http.ResponseWriter, r *http.Request) {
	uuid := r.URL.Path[len("/receipts/history/"):]
	if len(uuid) > 0 {
		handlerReceipe(uuid, w, r)
		return
	}

	js, bs, ds, err := getListOfFiles()
	if err != nil {
		handleError(w, r, err, 500, "Failed to load list of files")
		return
	}

	receiptFiles := loadFiles(js, bs, ds)

	data := processHistory(receiptFiles)

	err = buildHtmlHistory(w, data)
	if err != nil {
		handleError(w, r, err, 500, "Failed to render page")
		return
	}
}

func handlerReceipe(uuid string, w http.ResponseWriter, r *http.Request) {
	receiptResponse, err := readJsonFile(uuid)
	if err != nil {
		handleError(w, r, err, 500, fmt.Sprintf("Failed to read JSON file %s.json", uuid))
		return
	}

	sort.Sort(byTransactionId(receiptResponse.Receipt.InApp))

	receiptFile := &ReceiptFile{
		Uuid:            uuid,
		RequestDateMs:   receiptResponse.Receipt.RequestDateMs,
		ReceiptResponse: receiptResponse,
	}

	err = buildHtmlReceipt(w, receiptFile)
	if err != nil {
		handleError(w, r, err, 500, "Failed to render page")
		return
	}
}

func handlerTransactions(w http.ResponseWriter, r *http.Request) {
	js, bs, ds, err := getListOfFiles()
	if err != nil {
		handleError(w, r, err, 500, "Failed to load list of files")
		return
	}

	receiptFiles := loadFiles(js, bs, ds)

	data := processTransactions(receiptFiles)

	err = buildHtmlTransactions(w, data)
	if err != nil {
		handleError(w, r, err, 500, "Failed to render page")
		return
	}
}

func handleError(w http.ResponseWriter, r *http.Request, err error, code int, message string) {
	if err != nil {
		message = fmt.Sprintf("%s: %s", message, err)
	}

	log.Printf("%s %s %s %d %s: %v", r.RemoteAddr, r.Method, r.URL, code, message, err)

	errorBytes := []byte("test")

	w.WriteHeader(code)
	fmt.Fprintf(w, "%s", errorBytes)
}

func Log(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s %s", r.RemoteAddr, r.Method, r.URL)
		handler.ServeHTTP(w, r)
	})
}

func getListOfFiles() (map[string]bool, map[string]bool, map[string]bool, error) {
	bs := map[string]bool{}
	js := map[string]bool{}
	ds := map[string]bool{}

	files, err := ioutil.ReadDir(receiptsPath)
	if err != nil {
		return js, bs, ds, err
	}

	for _, file := range files {
		fileNameSlice := strings.Split(file.Name(), ".")
		uuid := fileNameSlice[0]

		if len(uuid) > 37 && uuid[36:] == "_debug" {
			ds[uuid[:36]] = true
			continue
		}

		switch fileNameSlice[1] {
		case "json":
			js[uuid] = true
			break
		case "base64":
			bs[uuid] = true
			break
		default:
			log.Printf("Found unsupported file %s.\n", file.Name())
		}
	}

	return js, bs, ds, nil
}

func loadFiles(jsonFiles map[string]bool, base64Files map[string]bool, debugJson map[string]bool) map[string]*ReceiptFile {
	result := map[string]*ReceiptFile{}

	for uuid, _ := range jsonFiles {
		receiptResponse, err := readJsonFile(uuid)
		if err != nil {
			log.Printf("Failed to read JSON file %s.json: %v", uuid, err)
			continue
		}

		hasBinary := false
		if _, ok := base64Files[uuid]; ok {
			hasBinary = true
		}

		hasDebugJson := false
		if _, ok := debugJson[uuid]; ok {
			hasDebugJson = true
		}

		result[uuid] = &ReceiptFile{
			Uuid:            uuid,
			RequestDateMs:   receiptResponse.Receipt.RequestDateMs,
			HasBinary:       hasBinary,
			HasDebugJson:    hasDebugJson,
			ReceiptResponse: receiptResponse,
		}
	}

	return result
}

func processTransactions(receiptFiles map[string]*ReceiptFile) *TemplateDataTransactions {
	transactions := map[string]InApp{}
	receipts := map[string]ReceiptFile{}

	for uuid, receiptFile := range receiptFiles {
		receipts[uuid] = ReceiptFile{
			Uuid:            uuid,
			RequestDateMs:   receiptFile.RequestDateMs,
			HasBinary:       receiptFile.HasBinary,
			ReceiptResponse: receiptFile.ReceiptResponse,
		}

		for _, inApp := range receiptFile.ReceiptResponse.Receipt.InApp {
			if _, ok := transactions[inApp.TransactionId]; ok {
				uuids := transactions[inApp.TransactionId].ReceiptUuids
				inApp.ReceiptUuids = append(uuids, uuid)
			} else {
				inApp.ReceiptUuids = []string{uuid}
			}

			transactions[inApp.TransactionId] = inApp
		}
	}

	for i, inApp := range transactions {
		if inApp.OriginalTransactionId != inApp.TransactionId {
			continue
		}

		inAppReceipts := filter(receipts, func(v ReceiptFile) bool {
			for _, uuid := range inApp.ReceiptUuids {
				if uuid == v.Uuid {
					return true
				}
			}

			return false
		})

		sort.Sort(byDate(inAppReceipts))
		inApp.ReceiptFile = inAppReceipts[0]

		transactions[i] = inApp
	}

	return &TemplateDataTransactions{
		InApp: transactions,
		// ReceiptFile: receipts,
	}
}

func processHistory(receiptFiles map[string]*ReceiptFile) *TemplateDataHistory {
	receipts := []ReceiptFile{}

	for uuid, receiptFile := range receiptFiles {
		receipts = append(
			receipts,
			ReceiptFile{
				Uuid:            uuid,
				RequestDateMs:   receiptFile.RequestDateMs,
				HasBinary:       receiptFile.HasBinary,
				HasDebugJson:    receiptFile.HasDebugJson,
				ReceiptResponse: receiptFile.ReceiptResponse,
			},
		)
	}

	sort.Sort(byDate(receipts))

	return &TemplateDataHistory{
		Receipts: receipts,
	}
}

func filter(vs map[string]ReceiptFile, f func(ReceiptFile) bool) []ReceiptFile {
	vsf := make([]ReceiptFile, 0)
	for _, v := range vs {
		if f(v) {
			vsf = append(vsf, v)
		}
	}
	return vsf
}

func readJsonFile(fileName string) (ReceiptResponse, error) {
	jsonBytes, err := ioutil.ReadFile(fmt.Sprintf("%s/%s.json", receiptsPath, fileName))
	if err != nil {
		return ReceiptResponse{}, err
	}

	receiptResponse := ReceiptResponse{}
	err = json.Unmarshal(jsonBytes, &receiptResponse)
	if err != nil {
		return ReceiptResponse{}, err
	}

	return receiptResponse, nil
}

func buildHtmlTransactions(w http.ResponseWriter, data *TemplateDataTransactions) error {
	t, err := template.ParseFiles("transactions.html")
	if err != nil {
		return err
	}

	err = t.Execute(w, data)
	if err != nil {
		return err
	}

	return nil
}

func buildHtmlHistory(w http.ResponseWriter, data *TemplateDataHistory) error {
	t, err := template.ParseFiles("history.html")
	if err != nil {
		return err
	}

	err = t.Execute(w, data)
	if err != nil {
		return err
	}

	return nil
}

func buildHtmlReceipt(w http.ResponseWriter, data *ReceiptFile) error {
	t, err := template.ParseFiles("receipt.html")
	if err != nil {
		return err
	}

	err = t.Execute(w, data)
	if err != nil {
		return err
	}

	return nil
}

func main() {
	log.Println("Starting...")

	port := os.Getenv("PORT")
	receiptsPath := os.Getenv("PATH_TO_RECEIPTS")

	log.Println("Using receipts path:", receiptsPath)

	http.Handle("/receipts/src/", http.StripPrefix("/receipts/src/", http.FileServer(http.Dir(receiptsPath))))
	http.HandleFunc("/receipts/history/", handlerHistory)
	http.HandleFunc("/receipts/transactions/", handlerTransactions)

	log.Printf("Listening on port %s...", port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", port), Log(http.DefaultServeMux)))
}
