
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
