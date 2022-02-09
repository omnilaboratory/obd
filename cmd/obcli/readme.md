## copy lndcli packge to init obcli packge
packages dir:
* lndcli :  ./lnd/cmd/lncli/*.go
* lndcli :  ./cmd/obcli

1. cp lncli files to obcli
```shell 
cp ./lnd/cmd/lncli/*.go ./cmd/obcli/
```
2. create file /cmd/obcli/main-obd.go 
3. cp main function from /cmd/obcli/main.go to main-obd.go 
4. modify  main-obd.go: add btcSubcmd and usdtsubcmd; assign lndci all  old-sumcommand（except createCommand and unlockCommand） to  btcSubcmd.Subcommands from line 71.   assign btcSubcmd and usdtsubcmd and createCommand and unlockCommand to app.Commands 
5. comment main function in the /cmd/obcli/main.go

now obcli app subcommand tree will look like below:
```shell
obcli
    create #new wallet
    unlock #unlock wallet
    btc #all-subcomand from old lndcli
        openchannel
        closechannel
        ...
        ...
        lnd-all-subcmd...
    usdt
      ...
      ...

```

## urfave to v2
urfave is the comand-line framework the project using,and the using verston v1 is old.

upgrade github.com/urfave/cli to v2: github.com/urfave/cli/v2

guide:
https://github.com/urfave/cli/blob/master/docs/migrate-v1-to-v2.md

now  we can use urfave all new freature, and can get latest support from
https://github.com/urfave/cli/issues
