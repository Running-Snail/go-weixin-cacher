package main

import (
    "flag"
    "log"
    "time"
    "zhihaojun.com/weixin"
    "github.com/bradfitz/gomemcache/memcache"
)

var mc *memcache.Client
var accessTokenKey string
var jssdkTicketKey string
var wx weixin.Weixin

func cacheAccessToken(accessTokenCH chan string) {
    log.Println("start to cache access token")
    for {
        accessTokenResp, err := wx.GetAccessToken()
        if err != nil {
            log.Println("error when get accessToken")
            log.Fatal(err)
            log.Println("try again in 1s")
            time.Sleep(time.Second)
            continue
        }
        if !accessTokenResp.Ok() {
            log.Println("accessTokenResp has error")
            log.Println("try again in 1s")
            time.Sleep(time.Second)
            continue
        }

        // write to storage
        log.Println("saving access token")
        err = mc.Set(&memcache.Item{
            Key: accessTokenKey,
            Object: map[string]interface{} {
                "access_token": accessTokenResp.AccessToken,
                "expires_in": accessTokenResp.ExpiresIn,
            },
        })
        if err != nil {
            log.Println("save access token error")
            log.Fatal(err)
        }

        accessTokenCH  <- accessTokenResp.AccessToken
        log.Printf("ready to sleep: %ds\n", accessTokenResp.ExpiresIn)
        time.Sleep(time.Duration(accessTokenResp.ExpiresIn) * time.Second)
    }
}

func cacheJSSDKTicket(accessTokenCH chan string) {
    log.Println("start to cache jssdk ticket")
    for {
        accessToken := <- accessTokenCH

        jssdkTicketResponse, err := wx.GetJSSDKTicket(accessToken)
        if err != nil {
            log.Println("error when get jssdk ticket")
            log.Fatal(err)
            log.Println("try again in 1s")
            time.Sleep(time.Second)
            continue
        }
        if !jssdkTicketResponse.Ok() {
            log.Println("jssdk ticket has error")
            log.Println("try again in 1s")
            time.Sleep(time.Second)
            continue
        }

        // write to storage
        log.Println("saving jssdk ticket")
        err = mc.Set(&memcache.Item{
            Key: jssdkTicketKey,
            Object: map[string]interface{} {
                "ticket": jssdkTicketResponse.Ticket,
                "expires_in": jssdkTicketResponse.ExpiresIn,
            },
        })
        if err != nil {
            log.Println("save jssdk ticket error")
            log.Fatal(err)
        }

        log.Printf("ready to sleep: %ds\n", jssdkTicketResponse.ExpiresIn)
        time.Sleep(time.Duration(jssdkTicketResponse.ExpiresIn) * time.Second)
    }
}

func main() {
    appId := flag.String("appid", "", "weixin appid")
    appSecret := flag.String("appsecret", "", "weixin appsecret")
    accessTokenKeyFlag := flag.String("a", "access_token", "access token key")
    jssdkTicketKeyFlag := flag.String("j", "jssdk_ticket", "jssdk ticket key")
    memcacheAddr := flag.String("mem_addr", "127.0.0.1:11211", "memcache addr")
    _ = flag.String("redis_addr", "127.0.0.1:6379", "redis addr")

    flag.Parse()

    accessTokenKey = *accessTokenKeyFlag
    jssdkTicketKey = *jssdkTicketKeyFlag

    log.Printf("appId=%s\n", *appId)
    log.Printf("appSecret=%s\n", *appSecret)
    log.Printf("accessTokenKey=%s\n", accessTokenKey)
    log.Printf("jssdkTicketKey=%s\n", jssdkTicketKey)
    log.Printf("memcacheAddr=%s\n", *memcacheAddr)

    accessTokenCH := make(chan string)
    wx = weixin.New(*appId, *appSecret)
    mc = memcache.New(*memcacheAddr)

    go cacheAccessToken(accessTokenCH)
    go cacheJSSDKTicket(accessTokenCH)

    for {
        time.Sleep(time.Second)
    }
}
