package pac

const pacRawTmpl = `

var methods = {
    direct: "DIRECT",
    blocked: "{{.BlockedMethod}}",
    default: "{{.DefaultMethod}}",
}

var hosts = {
    direct: [
        "",
        "{{.DirectDomains}}"
    ],
    blocked: [
        "{{.BlockedDomains}}"
    ]
}

var topLevel = {
    {{.TopLevel}}
};

var directAcc = {};
for (var i = 0; i < hosts.direct.length; i += 1) {
    directAcc[hosts.direct[i]] = true;
}

var blockedAcc = {};
for (var i = 0; i < hosts.blocked.length; i += 1) {
    blockedAcc[hosts.blocked[i]] = true;
}

// hostIsIP determines whether a host address is an IP address and whether
// it is private. Currenly only handles IPv4 addresses.
function hostIsIP(host) {
    var part = host.split('.');
    if (part.length != 4) {
        return [false, false];
    }
    var n;
    for (var i = 3; i >= 0; i--) {
        if (part[i].length === 0 || part[i].length > 3) {
            return [false, false];
        }
        n = Number(part[i]);
        if (isNaN(n) || n < 0 || n > 255) {
            return [false, false];
        }
    }
    if (part[0] == '127' || part[0] == '10' || (part[0] == '192' && part[1] == '168')) {
        return [true, true];
    }
    if (part[0] == '172') {
        n = Number(part[1]);
        if (16 <= n && n <= 31) {
            return [true, true];
        }
    }
    return [true, false];
}

function host2Domain(host) {
    var arr, isIP, isPrivate;
    arr = hostIsIP(host);
    isIP = arr[0];
    isPrivate = arr[1];
    if (isPrivate) {
        return "";
    }
    if (isIP) {
        return host;
    }

    var lastDot = host.lastIndexOf('.');
    if (lastDot === -1) {
        return ""; // simple host name has no domain
    }
    // Find the second last dot
    dot2ndLast = host.lastIndexOf(".", lastDot-1);
    if (dot2ndLast === -1)
        return host;

    var part = host.substring(dot2ndLast+1, lastDot);
    if (topLevel[part]) {
        var dot3rdLast = host.lastIndexOf(".", dot2ndLast-1);
        if (dot3rdLast === -1) {
            return host;
        }
        return host.substring(dot3rdLast+1);
    }
    return host.substring(dot2ndLast+1);
}

function FindProxyForURL(url, host) {
    if (url.substring(0,4) == "ftp:")
        return methods.direct;
    var domain = host2Domain(host);


    if (directAcc[host] || directAcc[domain]) {
        return methods.direct;
    } else if (blockedAcc[host] || blockedAcc[domain]) {
        return methods.blocked;
    } else
        return methods.default;

}
`
