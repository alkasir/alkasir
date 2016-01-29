// Tests

var testData, td, i;
var testsFailed = false;

testData = [
	{ ip: '127.0.0.1', isIP: true, isPrivate: true },
	{ ip: '127.2.1.1', isIP: true, isPrivate: true },
	{ ip: '192.168.1.1', isIP: true, isPrivate: true },
	{ ip: '172.16.1.1', isIP: true, isPrivate: true },
	{ ip: '172.20.1.1', isIP: true, isPrivate: true },
	{ ip: '172.31.1.1', isIP: true, isPrivate: true },
	{ ip: '172.15.1.1', isIP: true, isPrivate: false },
	{ ip: '172.32.1.1', isIP: true, isPrivate: false },
	{ ip: '10.16.1.1', isIP: true, isPrivate: true },
	{ ip: '12.3.4.5', isIP: true, isPrivate: false },
	{ ip: '1.2.3.4.5', isIP: false, isPrivate: false },
    { ip: 'google.com', isIP: false, isPrivate: false },
	{ ip: 'www.google.com.hk', isIP: false, isPrivate: false }
];

for (i = 0; i < testData.length; i += 1) {
	td = testData[i];
	arr = hostIsIP(td.ip);
	if (arr[0] !== td.isIP) {
		if (td.isIP) {
            testsFailed = true;
            console.log(td.ip + " is ip");
        } else {
            testsFailed = true;
            console.log(td.ip + " is NOT ip");
        }
    }
    if (arr[0] !== td.isIP) {
        if (td.isIP) {
            testsFailed = true;
            console.log(td.ip + " is private ip");
        } else {
            testsFailed = true;
            console.log(td.ip + " is NOT private ip");
        }
    }
}

testData = [
	// private ip should return direct
    { host: '192.168.1.1', mode: methods.direct},
    { host: '10.1.1.1', mode: methods.direct},
    { host: '172.16.2.1', mode: methods.direct},
    { host: '172.20.255.255', mode: methods.direct},
    { host: '172.31.255.255', mode: methods.direct},
    { host: '192.168.2.255', mode: methods.direct},

    // simple host should return methods.direct
    { host: 'localhost', mode: methods.direct},
    { host: 'simple', mode: methods.direct},

    // non private ip should return default
    { host: '172.15.0.255', mode: methods.default},
    { host: '12.20.2.1', mode: methods.default},

    // non private ip on blocked list should return blocked
    { host: '172.32.2.255', mode: methods.blocked},


    // host names
    { host: 'taobao.com', mode: methods.direct},
    { host: 'www.taobao.com', mode: methods.direct},
    { host: 'www.baidu.com', mode: methods.direct},
    { host: 'baidu.com', mode: methods.direct},
    { host: 'foo.baidu.com', mode: methods.direct},
    { host: 'google.com', mode: methods.default},
    { host: 'www.google.com', mode: methods.default},
    { host: 'www.google.com.hk', mode: methods.default},
    { host: 'alkasir.com', mode: methods.blocked},
    { host: 'some.domain', mode: methods.blocked},
    { host: 'www.some.domain', mode: methods.blocked}
];

for (i = 0; i < testData.length; i += 1) {
    td = testData[i];
    var res = FindProxyForURL("", td.host)
    if (res !== td.mode) {
        console.log(td.host + " should return " + td.mode + " but did return " + res);
        testsFailed = true ;
    }
}

if (testsFailed) {
    console.log("Tests failed!");
    process.exit(1);

}
