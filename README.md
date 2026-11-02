# ipset
ipset plugin for CoreDNS

[中文简要使用说明](https://github.com/missdeer/ipset/issues/4#issuecomment-974837810)，感谢[@ioiioo](https://github.com/ioiioo)。

## Example

```js
.:53 {
        forward . 192.168.1.1
        ipset {
                twiroute api.twitter.com twitter.com dev.twitter.com www.twitter.com
                route2 domain1.net domain2.net domain3.net
        }
        ipset testroute1 testroute2 testroute3
        log
}
```

On macOS, ipset writes directly to `/dev/pf` through `github.com/leiless/dnsredir/pf`.
Run CoreDNS with permission to open `/dev/pf`; building requires CGO and the PF headers
described in dnsredir's `pf/download-pf-headers.sh`.
Use `NAME:ANCHOR` to select an anchored table, for example:

```text
ipset {
    direct_cidr:com.apple/climbovergui example.com
}
```

Tables must already exist. Duplicate addresses are handled by PF, and shutdown does
not flush tables owned by other applications.
