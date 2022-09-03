package main

import(
    "bytes"
    "fmt"
    //"reflect"
    cbugorji "github.com/ugorji/go/codec"
)

type Logger struct {
    Vin                     string
    State                   uint64
    VehicleID               string
    DeviceID                string
    LoggerID                uint64 // change to primitive.ObjectID
    IntegratorAccountID     string
    Timestamp               uint64
    Items                   map[string][]LoggerItem
    Model                   string
}

type LoggerItem struct {
    ItemValue       interface{}
    ItemTime        uint64
}

type Location struct {
    Latitude        int64
    Longitude       int64
}

var itemIdToString = map[string]uint64 {
    "VehicleSpeed":     288,
    "BatteryLevel":     1600,
    "GNSS":             2048,
}

func decodeCbor(in []byte) (interface{}, error) { 
    var returnedValue interface{}
    bufferCBOR := bytes.NewBuffer(in)
    err := cbugorji.NewDecoderBytes(bufferCBOR.Bytes(), new(cbugorji.CborHandle)).Decode(&returnedValue)
    return returnedValue, err
}

func getValue(data interface{}) (int64) {
    var intValue int64

    switch v := data.(type) {
    case int64:
        intValue = v
    case uint64:
        diff := v
        intValue = int64(diff)
    }

    return intValue
}

func GetItemNameFromItemID(value uint64) (name string) {
    for k, v := range itemIdToString {
        if v == value {
            name = k
            return
        }
    }
    return
}

func parsingLoggerDataCbor(stream interface{}) {
    
    // Getting records
    for _, recordData := range stream.([]interface{}) {
        record := recordData.([]interface{})

        header, diff := record[0], record[1:]

        //Parsing Header
        headerValues := header.([]interface{})
        state := headerValues[0].(uint64)
        indexes := headerValues[1]
        //timestamp := headerValues[2]
        initialValues := headerValues[3:]

        recordDataLogger := Logger {
            Vin:                    "",
            State:                  state,
            VehicleID:              "", 
            DeviceID:               "",
            LoggerID:               85,
            IntegratorAccountID:    "",
            Timestamp:              001,
            Items:                  make(map[string][]LoggerItem),
            Model:                  "",
        }

        // Getting the initial values
        for idx, id := range indexes.([]interface{}) {

            key := GetItemNameFromItemID(id.(uint64))
            loggerData := LoggerItem {
                ItemTime:   headerValues[2].(uint64),
            }

            if id.(uint64) == 2048 {
                loggerData.ItemValue = Location {
                    Longitude:      getValue(initialValues[idx].([]interface{})[0]),
                    Latitude:       getValue(initialValues[idx].([]interface{})[1]),
                } 
            } else {
                loggerData.ItemValue = getValue(initialValues[idx])
            }
            recordDataLogger.Items[key] = append(recordDataLogger.Items[key],  loggerData)
        }

        // Calculating current values
        for _, sample := range diff {
            itemSample := sample.([]interface{})
            Id := indexes.([]interface{})[itemSample[1].(uint64)].(uint64)
            itemKey := GetItemNameFromItemID(Id)

            previousData := recordDataLogger.Items[itemKey][len(recordDataLogger.Items[itemKey]) - 1]
            itemData := LoggerItem {
                ItemTime:   previousData.ItemTime + itemSample[0].(uint64),
            }

            if Id == 2048 {
                itemData.ItemValue = Location {
                    Longitude:      previousData.ItemValue.(Location).Longitude + getValue(itemSample[2].([]interface{})[0]),
                    Latitude:       previousData.ItemValue.(Location).Latitude + getValue(itemSample[2].([]interface{})[1]),
                }
            } else {
                itemData.ItemValue = previousData.ItemValue.(int64) + getValue(itemSample[2])
            }

            recordDataLogger.Items[itemKey] = append(recordDataLogger.Items[itemKey], itemData)


        }
        fmt.Println(recordDataLogger)
    }
}

func main() {
    
    cbor := []byte{ 
        0x81, 0x97, 0x86, 0x03, 0x83, 0x19, 0x08, 0x00, 0x19, 0x01, 0x20, 0x19, 0x06, 0x40, 0x1A, 0x62, 0xD8, 0xF0, 
        0x0B, 0x82, 0x19, 0xCB, 0x8C, 0x19, 0xC8, 0xB0, 0x18, 0x82, 0x0C, 0x83, 0x01, 0x00, 0x82, 0x0A, 0x18, 0x64, 
        0x83, 0x01, 0x01, 0x24, 0x83, 0x01, 0x00, 0x82, 0x39, 0x01, 0x17, 0x19, 0x02, 0x39, 0x83, 0x01, 0x01, 0x02, 
        0x83, 0x01, 0x00, 0x82, 0x04, 0x29, 0x83, 0x01, 0x01, 0x0A, 0x83, 0x01, 0x00, 0x82, 0x00, 0x00, 0x83, 0x01, 
        0x01, 0x33, 0x83, 0x01, 0x00, 0x82, 0x39, 0x07, 0x73, 0x39, 0x06, 0x20, 0x83, 0x01, 0x01, 0x04, 0x83, 0x01, 
        0x02, 0x20, 0x83, 0x01, 0x00, 0x82, 0x0A, 0x18, 0x64, 0x83, 0x01, 0x01, 0x24, 0x83, 0x01, 0x00, 0x82, 0x39, 
        0x01, 0x17, 0x19, 0x02, 0x39, 0x83, 0x01, 0x01, 0x02, 0x83, 0x01, 0x00, 0x82, 0x04, 0x29, 0x83, 0x01, 0x01, 
        0x0A, 0x83, 0x01, 0x00, 0x82, 0x00, 0x00, 0x83, 0x01, 0x01, 0x33, 0x83, 0x01, 0x00, 0x82, 0x39, 0x07, 0x73, 
        0x39, 0x06, 0x20, 0x83, 0x01, 0x01, 0x04, 0x83, 0x01, 0x02, 0x20,
    }

    /*cbor := []byte{
        0x82, 0x8C, 0x86, 0x03, 0x83, 0x19, 0x08, 0x00, 0x19, 0x01, 0x20, 0x19, 0x06, 0x40, 0x1A, 0x62, 0xD8, 0xF0,
        0x0B, 0x82, 0x19, 0xCB, 0x8C, 0x19, 0xC8, 0xB0, 0x18, 0x82, 0x0C, 0x83, 0x01, 0x00, 0x82, 0x0A, 0x18, 0x64,
        0x83, 0x01, 0x01, 0x24, 0x83, 0x01, 0x00, 0x82, 0x39, 0x01, 0x17, 0x19, 0x02, 0x39, 0x83, 0x01, 0x01, 0x02,
        0x83, 0x01, 0x00, 0x82, 0x04, 0x29, 0x83, 0x01, 0x01, 0x0A, 0x83, 0x01, 0x00, 0x82, 0x00, 0x00, 0x83, 0x01,
        0x01, 0x33, 0x83, 0x01, 0x00, 0x82, 0x39, 0x07, 0x73, 0x39, 0x06, 0x20, 0x83, 0x01, 0x01, 0x04, 0x83, 0x01, 
        0x02, 0x20, 0x87, 0x85, 0x04, 0x82, 0x19, 0x08, 0x00, 0x19, 0x06, 0x40, 0x1A, 0x62, 0xD8, 0xF0, 0x10, 0x82,
        0x19, 0xCB, 0x8C, 0x19, 0xC8, 0xB0, 0x0C, 0x83, 0x01, 0x00, 0x82, 0x0A, 0x18, 0x64, 0x83, 0x01, 0x00, 0x82,
        0x39, 0x01, 0x17, 0x19, 0x02, 0x39, 0x83, 0x01, 0x00, 0x82, 0x04, 0x29, 0x83, 0x01, 0x00, 0x82, 0x00, 0x00,
        0x83, 0x01, 0x00, 0x82, 0x39, 0x07, 0x73, 0x39, 0x06, 0x20, 0x83, 0x01, 0x01, 0x20,
    }*/
    
    decodedCbor, err := decodeCbor(cbor)

    if err != nil {
        fmt.Println(err)
    }

    parsingLoggerDataCbor(decodedCbor)
}
