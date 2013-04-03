"use strict";

module.exports = {
    obfuscate: obfuscate,
    unveil: unveil
};

var _ = require('underscore')._;
var async = require('async');
var uuid = require('node-uuid');

var encryptAttrs = ['password', 'key', 'Xauth_password', 'IPSec_secret'];
var encryptTag = '$mole$';

function obfuscateObj(obj, keyStore) {
    for (var i = 0; i < encryptAttrs.length; i++) {
        var attr = encryptAttrs[i];
        if (obj[attr] && obj[attr].indexOf(encryptTag) !== 0) {
            var u = uuid.v4();
            keyStore.set(u, obj[attr]);
            obj[attr] = encryptTag + u;
        }
    }
    return obj;
}

function unveilObj(obj, client, cb) {
    var operations = [];
    encryptAttrs.forEach(function (attr) {
        if (obj[attr] && obj[attr].indexOf(encryptTag) === 0) {
            operations.push(function (cb) {
                var key = obj[attr].substr(encryptTag.length);
                client.getKey(key, function (unveiled) {
                    obj[attr] = unveiled;
                    cb(null);
                })
            })
        }
    });

    async.series(operations, function () {
        cb(obj);
    })
}

function obfuscate(obj, keyStore) {
    if (obj.hosts) {
        for (var hostname in obj.hosts) {
            obj.hosts[hostname] = obfuscateObj(obj.hosts[hostname], keyStore);
        }
    }
    if (obj.vpnc)
        obj.vpnc = obfuscateObj(obj.vpnc, keyStore);
    if (obj.openconnect)
        obj.openconnect = obfuscateObj(obj.openconnect, keyStore);
    return obj;
}

function unveil(obj, client, cb) {
    var operations = [];
    if (obj.hosts) {
        Object.keys(obj.hosts).forEach(function (hostname) {
            operations.push(function (cb) {
                unveilObj(obj.hosts[hostname], client, function (unveiled) {
                    obj.hosts[hostname] = unveiled;
                    cb(null);
                });
            });
        });
    }

    if (obj.vpnc) {
        operations.push(function (cb) {
            unveilObj(obj.vpnc, client, function (unveiled) {
                obj.vpnc = unveiled;
                cb(null);
            });
        });
    }

    if (obj.openconnect) {
        operations.push(function (cb) {
            unveilObj(obj.openconnect, client, function (unveiled) {
                obj.openconnect = unveiled;
                cb(null);
            });
        });
    }

    async.series(operations, function () {
        console.dir(obj)
        cb(obj);
    });
}
