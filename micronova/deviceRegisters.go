package micronova

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"slices"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
)

const (
	deviceRegistersPath = "deviceGetRegistersMap"
	ValueLanguage       = "ENG"
)

type deviceRegistersReq struct {
	ProductId  string `json:"id_product"`
	DeviceId   string `json:"id_device"`
	LastUpdate string `json:"last_update"`
}

type encVal struct {
	Value       int    `json:"value"`
	Lang        string `json:"lang"`
	Description string `json:"description"`
}

type register struct {
	AppImageName     string   `json:"app_image_name"`
	RegKey           string   `json:"reg_key"`
	RegName          string   `json:"reg_name"`
	RegNameApp       string   `json:"reg_name_app"`
	RegKeyGroup      string   `json:"reg_key_group"`
	RegType          string   `json:"reg_type"`
	Offset           int      `json:"offset"`
	Eeprom           bool     `json:"eeprom"`
	IsWord           bool     `json:"is_word"`
	BigEndian        bool     `json:"big_endian"`
	Formula          string   `json:"formula"`
	FormulaInverse   string   `json:"formula_inverse"`
	FormatString     string   `json:"format_string"`
	SetMin           int      `json:"set_min"`
	SetMax           int      `json:"set_max"`
	Readonly         bool     `json:"readonly"`
	Mask             int      `json:"mask"`
	IsHex            bool     `json:"is_hex"`
	IsTemperature    bool     `json:"is_temperature"`
	UseApp           bool     `json:"use_app"`
	DisplayDashboard bool     `json:"display_dashboard"`
	DisplayManage    bool     `json:"display_manage"`
	DisplayInfo      bool     `json:"display_info"`
	UseCat           bool     `json:"use_cat"`
	DisplayOrder     int      `json:"display_order"`
	Step             int      `json:"step"`
	UseHistory       bool     `json:"use_history"`
	UseHistoryOutput bool     `json:"use_history_output"`
	WithSign         bool     `json:"with_sign"`
	FromCharCode     bool     `json:"from_char_code"`
	NotifyOnChange   bool     `json:"notify_on_change"`
	TabGroup         string   `json:"tab_group"`
	IsTest           bool     `json:"is_test"`
	EncVal           []encVal `json:"enc_val"`
}

type registersMap struct {
	Id           string     `json:"id"`
	CustumerCode string     `json:"customer_code"`
	CreationDate customTime `json:"creation_date"`
	LastUpdate   customTime `json:"last_update"`
	Registers    []register `json:"registers"`
}

type deviceRegistersMap struct {
	RegistersMap []registersMap `json:"registers_map"`
}

type deviceRegistersResp struct {
	Success            bool               `json:"Success"`
	Text               string             `json:"Text"`
	Value              bool               `json:"Value"`
	DeviceRegistersMap deviceRegistersMap `json:"device_registers_map"`
	Error              string             `json:"Error"`
}

type customTime time.Time

func (ct *customTime) UnmarshalJSON(b []byte) error {
	// Trim the quotes from the JSON string
	str := strings.Trim(string(b), `"`)

	// Parse the time using the desired layout (e.g., RFC 3339)
	t, err := time.Parse("2006-01-02T15:04:05", str)
	if err != nil {
		return err
	}

	*ct = customTime(t) // Assign the parsed time to the custom type
	return nil
}

func deviceRegisters() {
	if len(offsets) == 0 && len(dm.Config.Micronova.RegKeys) == 0 {
		log.Error().Msg("unable to get deviceRegisters, offsets and RegKeys are empty")
		return
	}

	// Create context for timeout
	ctx, cancel := context.WithTimeout(context.Background(), requestTimeout)
	defer cancel()

	req := buildDeviceRegistersRequest(ctx)
	if req == nil {
		log.Fatal().Msg("Failed to build DeviceRegisters request")
	}

	log.Info().Msgf("API Call: %s", deviceRegistersPath)

	httpClient := &http.Client{Timeout: requestTimeout}
	resp, err := httpClient.Do(req)
	if err != nil {
		log.Fatal().Err(err).Msg("DeviceRegisters HTTP Client error")
		return
	}
	defer resp.Body.Close()

	var result deviceRegistersResp
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		log.Fatal().Err(err).Msg("Error decoding response")
	}

	//log.Trace().Msgf("DeviceRegisters Response:\n%+v", result)

	// Validate HTTP status code and response Success field
	if resp.StatusCode != http.StatusOK || !result.Success {
		log.Fatal().Msgf("DeviceRegisters Unsuccesful (%s) %s %s", resp.Status, result.Error, result.Text)
	}

	if len(result.DeviceRegistersMap.RegistersMap) == 0 {
		log.Fatal().Msg("DeviceRegisters: No DeviceRegistersMap")
	}

	for _, reg := range result.DeviceRegistersMap.RegistersMap[0].Registers {
		var selected bool
		var topicKey string
		if len(offsets) != 0 {
			selected = slices.Contains(offsets, reg.Offset)
			topicKey = reg.RegKey
		}
		if len(dm.Config.Micronova.RegKeys) != 0 {
			for _, regKey := range dm.Config.Micronova.RegKeys {
				if regKey.Key == reg.RegKey {
					selected = true
					topicKey = regKey.Topic
					break
				}
			}
		}
		if selected && (reg.RegType == "SET" || reg.RegType == "GET") {
			param := parameter{
				regKey:   reg.RegKey,
				topicKey: topicKey,
				offset:   reg.Offset,
				mask:     reg.Mask,
				minimum:  reg.SetMin,
				maximum:  reg.SetMax,
				formula:  reg.Formula,
				format:   reg.FormatString,
			}
			if len(reg.EncVal) != 0 {
				for _, encval := range reg.EncVal {
					if encval.Lang == ValueLanguage {
						valueDescr := valueDescr{encval.Value, encval.Description}
						param.valueDescr = append(param.valueDescr, valueDescr)
					}
				}
			}
			parameters = append(parameters, param)
		}
	}

	for _, par := range parameters {
		log.Trace().Msgf("Parameter %s: %v", par.regKey, par)
	}
}

func buildDeviceRegistersRequest(ctx context.Context) *http.Request {
	registers := deviceRegistersReq{
		ProductId:  dm.Session.ProductId,
		DeviceId:   dm.Session.DeviceId,
		LastUpdate: "2018-06-03T08:59:54.043",
	}

	// Marshal JSON for request
	jsonData, err := json.Marshal(registers)
	if err != nil {
		log.Error().Err(err).Msg("Failed to marshal request")
		return nil
	}

	reqUrl := url.URL{Scheme: "https", Host: apiDomain, Path: deviceRegistersPath}
	payload := bytes.NewReader(jsonData)
	log.Trace().Msgf("DeviceRegisters Request:\n%v", string(jsonData))

	// Create Request
	req, err := http.NewRequestWithContext(ctx, "POST", reqUrl.String(), payload)
	if err != nil {
		log.Error().Err(err).Msg("HTTP Request error")
		return nil
	}

	// Set HTTP headers
	req.Header.Set("Accept", acceptHeader)
	req.Header.Set("Content-Type", contentTypeHeader)
	req.Header.Set("id_brand", brandIdHeader)
	req.Header.Set("customer_code", customerCode)
	req.Header.Set("local", "false")
	req.Header.Set("Authorization", dm.Session.Token)

	return req
}
