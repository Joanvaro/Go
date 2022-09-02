package main

import(
    "bytes"
    "fmt"
    //"reflect"
    cbugorji "github.com/ugorji/go/codec"
)

type Event struct {
    Time    uint64
    State   uint64
    Items   []DataLogger
}

type DataLogger struct {
    Id      uint64
    Time    uint64
    Value   []Value
}

type Value struct {
    Time    uint64
    Value   []interface{}
}

type Logger struct {
    Vin                     string
    State                   uint64
    VehicleID               string
    DeviceID                string
    LoggerID                uint64 // change to primitive.ObjectID
    IntegratorAccountID     string
    Timestamp               uint64
    Items                   map[string]LoggerItem
    Model                   string
}

type LoggerItem struct {
    ItemValue       interface{}
    ItemTime        uint64
}

var itemIdToString = map[int]string {
    288:    "VehicleSpeed",
    1600:   "BatteryLevel",
    2048:   "GNSS",
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

func parsingHeader(header interface{}) (Event){

    initialValues := header.([]interface{})

    rawEvent := Event {
        State:      initialValues[0].(uint64), 
        Time:       initialValues[2].(uint64),
    }

    for idx, id := range initialValues[1].([]interface{}) {
        item := DataLogger {
            Time:   initialValues[2].(uint64),
            Id:     id.(uint64),
        } 

        initialValue := Value {
            Time:   initialValues[2].(uint64),
        }

        if item.Id == 2048 {
            GpsData := []int64{}
            
            for _, val := range initialValues[3+idx].([]interface{}) {
                value := getValue(val)
                GpsData = append(GpsData, value)
            }

            initialValue.Value = append(initialValue.Value, GpsData)
        } else {
            newValue := getValue(initialValues[3+idx])
            initialValue.Value = append(initialValue.Value, newValue)
        }

        item.Value = append(item.Value, initialValue)
        rawEvent.Items = append(rawEvent.Items, item)
    }
    return rawEvent
}

func parsingDiff(diff []interface{}, events Event) (Event){

    for _, sample := range diff {
        sample := sample.([]interface{})
        index := sample[1].(uint64)
        
        currentValue := Value{}

        if events.Items[index].Id == 2048 {
            GpsData := []int64{}

            lastValue := events.Items[index].Value[len(events.Items[index].Value) - 1]

            for idx, val := range sample[2].([]interface{}) {
                newValue := lastValue.Value[0].([]int64)[idx] + getValue(val)
                GpsData = append(GpsData, newValue)
            }
            currentValue.Time = lastValue.Time + sample[0].(uint64)
            currentValue.Value = append(currentValue.Value, GpsData)
            events.Items[index].Value = append(events.Items[index].Value, currentValue)
        } else {

            lastValue := events.Items[index].Value[len(events.Items[index].Value) - 1]
            newValue := lastValue.Value[0].(int64) + getValue(sample[2])

            currentValue.Time = lastValue.Time + sample[0].(uint64)
            currentValue.Value = append(currentValue.Value, newValue)
            events.Items[index].Value = append(events.Items[index].Value, currentValue)
        }
    }
    return events
}

func parsingStream(stream interface{})([]Event) {

    var events []Event

    for _, rcrd := range stream.([]interface{}) {
        var record []interface{}
        record = rcrd.([]interface{})

        header, diff := record[0], record[1:]
        fmt.Println("header = ", header)
        fmt.Println("diff = ", diff)

        streamEvents := parsingHeader(header) 
        event := parsingDiff(diff, streamEvents)

        events = append(events, event)
    }

    return events
}

func fillingLoggerStruct(events []Event) {

    for _, record := range events {
    
        recordStream := Logger {
            Vin:                    "1FRE4GB67JUKW34DV",
            State:                  record.State,
            VehicleID:              "vfg45312",
            DeviceID:               "1225486",
            IntegratorAccountID:    "A1548FG",
            Model:                  "Versa",
        }

        for _, item := range record.Items {
            fmt.Println(item)
            fmt.Println("")
        }

        fmt.Println("")
        fmt.Println(recordStream)
    }


}

func main() {
    
    cbor := []byte{ 
        0x81, 0x97, 0x86, 0x03, 0x83, 0x19, 0x08, 0x00, 0x19, 0x01, 0x20, 0x19, 0x06, 0x40, 0x1A, 0x62, 0xD8, 0xF0, 0x0B,
        0x82, 0x19, 0xCB, 0x8C, 0x19, 0xC8, 0xB0, 0x18, 0x82, 0x0C, 0x83, 0x01, 0x00, 0x82, 0x0A, 0x18, 0x64, 0x83,
        0x01, 0x01, 0x24, 0x83, 0x01, 0x00, 0x82, 0x39, 0x01, 0x17, 0x19, 0x02, 0x39, 0x83, 0x01, 0x01, 0x02, 0x83,
        0x01, 0x00, 0x82, 0x04, 0x29, 0x83, 0x01, 0x01, 0x0A, 0x83, 0x01, 0x00, 0x82, 0x00, 0x00, 0x83, 0x01, 0x01,
        0x33, 0x83, 0x01, 0x00, 0x82, 0x39, 0x07, 0x73, 0x39, 0x06, 0x20, 0x83, 0x01, 0x01, 0x04, 0x83, 0x01, 0x02,
        0x20, 0x83, 0x01, 0x00, 0x82, 0x0A, 0x18, 0x64, 0x83, 0x01, 0x01, 0x24, 0x83, 0x01, 0x00, 0x82, 0x39, 0x01,
        0x17, 0x19, 0x02, 0x39, 0x83, 0x01, 0x01, 0x02, 0x83, 0x01, 0x00, 0x82, 0x04, 0x29, 0x83, 0x01, 0x01, 0x0A,
        0x83, 0x01, 0x00, 0x82, 0x00, 0x00, 0x83, 0x01, 0x01, 0x33, 0x83, 0x01, 0x00, 0x82, 0x39, 0x07, 0x73, 0x39,
        0x06, 0x20, 0x83, 0x01, 0x01, 0x04, 0x83, 0x01, 0x02, 0x20,
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

    events := parsingStream(decodedCbor)
    fillingLoggerStruct(events)
}
