# Exchanges

## Development requirements

```
go get github.com/golang/mock/mockgen
go install github.com/princjef/gomarkdoc/cmd/gomarkdoc@latest
go install github.com/russross/blackfriday-tool@latest
```

## Notes

For Binance/Long we use GTC+Limit orders

Documentation for utils are stored in `utils/doc.md` and `utils/doc.html`.

## Useful commands

```
bash generate.sh
websocat 'wss://stream.binance.com:9443/ws/!miniTicker@arr'

go get github.com/princjef/gomarkdoc/cmd/gomarkdoc # for Markdown documentation generation
go get github.com/russross/blackfriday-tool # for Markdown to HTML conversion

bash -c "while [ 1 ]; do ./show_data orders -symbol='LTCUSDT' -from-day='21-01-26'; sleep 5; done"
bash -c "while [ 1 ]; do ./show_data wsprice -symbol='LTCUSDT'; sleep 2; done"
bash -c "while [ 1 ]; do ./show_data wsorders -symbol='LTCUSDT'; sleep 2; done"
bash -c "while [ 1 ]; do ./show_data position; sleep 2; done"
bash -c "while [ 1 ]; do ./show_data getsymbols; sleep 2; done"
bash -c "while [ 1 ]; do ./show_data account 2>&1 | grep -E 'USDT|LTC'; sleep 1; done"
bash -c "while [ 1 ]; do ./show_data account 2>&1; sleep 1; done"
bash -c "while [ 1 ]; do ./show_data createorder -symbol='ETHUSDT'; sleep 2; done"
bash -c "while [ 1 ]; do ./show_data cancel -symbol='ETHUSDT' -coid='PCR8OyZ6qOTgqPldZgqQBL'; sleep 2; done"

websocat 'wss://stream.binance.com:9443/ws/ltcusdt@miniTicker' | jq "{s:.s,price:.c}"

./show_data cancel -symbol=LTCUSDT -coid="RUN200-573f0701"
```
