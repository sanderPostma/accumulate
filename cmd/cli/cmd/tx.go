package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/AccumulateNetwork/accumulated/internal/url"
	"github.com/AccumulateNetwork/accumulated/types"
	acmeapi "github.com/AccumulateNetwork/accumulated/types/api"
	"github.com/spf13/cobra"
)

var txCmd = &cobra.Command{
	Use:   "tx",
	Short: "Create and get token txs",
	Run: func(cmd *cobra.Command, args []string) {

		if len(args) > 0 {
			switch arg := args[0]; arg {
			case "get":
				if len(args) > 1 {
					GetTX(args[1])
				} else {
					fmt.Println("Usage:")
					PrintTXGet()
				}
			case "history":
				if len(args) > 3 {
					GetTXHistory(args[1], args[2], args[3])
				} else {
					fmt.Println("Usage:")
					PrintTXHistoryGet()
				}
			case "create":
				if len(args) > 3 {
					CreateTX(args[1], args[2:])
				} else {
					fmt.Println("Usage:")
					PrintTXCreate()
				}
			default:
				fmt.Println("Usage:")
				PrintTX()
			}
		} else {
			fmt.Println("Usage:")
			PrintTX()
		}

	},
}

func init() {
	rootCmd.AddCommand(txCmd)
}

func PrintTXGet() {
	fmt.Println("  accumulate tx get [txid]			Get token transaction by txid")
}

func PrintTXCreate() {
	fmt.Println("  accumulate tx create [from] [to] [amount]	Create new token tx")
}

func PrintTXHistoryGet() {
	fmt.Println("  accumulate tx history [url] [starting transaction number] [ending transaction number]	Get transaction history")
}

func PrintTX() {
	PrintTXGet()
	PrintTXCreate()
	PrintTXHistoryGet()
}

func GetTX(hash string) {

	var res interface{}
	var str []byte
	var hashbytes types.Bytes32

	params := new(acmeapi.TokenTxRequest)
	err := hashbytes.FromString(hash)
	if err != nil {
		log.Fatal(err)
	}

	params.Hash = hashbytes

	data, err := json.Marshal(params)
	jsondata := json.RawMessage(data)
	if err != nil {
		log.Fatal(err)
	}

	if err := Client.Request(context.Background(), "token-tx", jsondata, &res); err != nil {
		log.Fatal(err)
	}

	str, err = json.Marshal(res)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(string(str))

}

func GetTXHistory(accountUrl string, s string, e string) {

	start, err := strconv.Atoi(s)
	if err != nil {
		log.Fatal(err)
	}
	end, err := strconv.Atoi(e)
	if err != nil {
		log.Fatal(err)
	}

	u, err := url.Parse(accountUrl)
	if err != nil {
		log.Fatal(err)
	}

	var res interface{}
	var str []byte

	params := new(acmeapi.APIRequestURLPagination)
	params.URL = types.String(u.String())
	params.Start = int64(start)
	params.Limit = int64(end)

	data, err := json.Marshal(params)
	jsondata := json.RawMessage(data)
	if err != nil {
		log.Fatal(err)
	}

	if err := Client.Request(context.Background(), "token-account-history", jsondata, &res); err != nil {
		log.Fatal(err)
	}

	str, err = json.Marshal(res)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(string(str))

}

func CreateTX(sender string, args []string) {
	//sender string, receiver string, amount string
	var res interface{}
	var str []byte
	var err error
	u, err := url.Parse(sender)
	if err != nil {
		log.Fatal(err)
	}

	args, si, pk, err := prepareSigner(u, args)

	if len(args) < 2 {
		PrintTXCreate()
		log.Fatal("invalid number of arguments for tx create")
	}

	u2, err := url.Parse(args[0])
	if err != nil {
		log.Fatalf("invalid receiver url %s, %v", args[0], err)
	}
	amount := args[1]

	//fmt.Println(hex.EncodeToString(pk))
	tokentx := new(acmeapi.TokenTx)
	tokentx.From = types.UrlChain{types.String(u.String())}

	to := []*acmeapi.TokenTxOutput{}
	r := &acmeapi.TokenTxOutput{}
	amt, err := strconv.ParseUint(amount, 10, 64)
	r.Amount = uint64(amt)
	r.URL.String = types.String(u2.String())
	to = append(to, r)
	tokentx.To = to

	data, err := json.Marshal(tokentx)
	if err != nil {
		log.Fatal(err)
	}

	dataBinary, err := tokentx.MarshalBinary()
	if err != nil {
		log.Fatal(err)
	}

	nonce := uint64(time.Now().Unix())
	params, err := prepareGenTx(data, dataBinary, u, si, pk, nonce)

	if err := Client.Request(context.Background(), "token-tx-create", params, &res); err != nil {
		log.Fatal(err)
	}

	str, err = json.Marshal(res)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(string(str))

	if err != nil {
		log.Fatal(err)
	}

}
