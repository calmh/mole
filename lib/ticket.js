"use strict";

var crypto = require('crypto');

var key = crypto.randomBytes(64).toString('base64');
var hashAlgo = 'sha1';
var cipherAlgo = 'aes256';

exports.issue = function (user, ip, validity) {
    if (!key) {
        return null;
    }

    var hf = crypto.createHash(hashAlgo);

    var tdata = [crypto.randomBytes(16).toString('hex'), user, ip, Math.round(Date.now() / 1000) + validity].join(";");
    var hash = hf.update(tdata).digest('base64');
    var ticket = [tdata, hash].join("&");

    var cf = crypto.createCipher(cipherAlgo, key);
    return Buffer.concat([cf.update(ticket, 'utf8'), cf.final()]).toString('base64');
};

exports.examine = function (ticket) {
    if (!key) {
        return null;
    }

    try {
        var cf = crypto.createDecipher(cipherAlgo, key);
        var msg = Buffer.concat([cf.update(ticket, 'base64'), cf.final()]).toString('utf8');

        var parts = msg.split("&");
        var data = parts[0];
        var hash = parts[1];

        var hf = crypto.createHash(hashAlgo);
        var curHash = hf.update(data).digest('base64');
        if (hash !== curHash) {
            return null;
        }

        parts = data.split(";");
        if (parts.length !== 4) {
            return null;
        }

        var obj = {
            user: parts[1],
            ip: parts[2],
            valid: parseInt(parts[3], 10),
            nonce: parts[0],
        };

        if (obj.valid < Date.now() / 1000) {
            return null;
        }

        return obj;
    } catch (e) {
        return null;
    }
};
