

var methods = {
    default: "DEFAULT",
    direct: "DIRECT",
    blocked: "SOCKS5 127.0.0.1:4488",
}

var hosts = {
    direct: [
        "", // corresponds to simple host name and ip address
        "taobao.com",
        "www.baidu.com",
        "baidu.com"
    ],
    blocked: [
        "alkasir.com",
        "some.domain",
        "172.32.2.255",
    ],
}

var topLevel = {
    "ac": true,
    "co": true,
    "com": true,
    "edu": true,
    "gov": true,
    "net": true,
    "se": true,
    "org": true
};
