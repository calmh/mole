#!/usr/bin/env node

var _ = require('underscore');
var fs = require('fs');
var https = require('https');

// FIXME: We should not use the con functions
var con = require('./console');

var defaults = {
    host: undefined,
    port: undefined,
    key: undefined,
    cert: undefined,
};

module.exports = {
    init: init,
    list: serverList,
    send: serverSend,
    fetch: serverFetch,
    register: serverRegister,
    token: serverToken,
    newUser: serverNewUser,
    delUser: serverDelUser,
    saveBin: serverSaveBinary,
}

function init(options) {
    _.extend(defaults, options);
}

function unbufferedRequest(options, callback) {
    _.defaults(options, defaults);
    _.defaults(options, { method: 'GET', agent: false });

    var req = https.request(options, function (res) {
        var serverCert = res.connection.getPeerCertificate();
        var serverFingerprint = serverCert.fingerprint;

        if (res.statusCode !== 200) {
            con.fatal('Server returned code ' + res.statusCode);
        }

        if (options.fingerprint && options.fingerprint !== serverFingerprint) {
            con.fatal('Presented server certificate does not match stored fingerprint');
        }

        callback(res);
    });

    req.on('error', function (err) {
        con.fatal(err);
    });

    return req;
}

function bufferedRequest(options, callback) {
    var buffer = '';
    var req = unbufferedRequest(options, function (res) {
        var serverCert = res.connection.getPeerCertificate();
        var serverFingerprint = serverCert.fingerprint;

        res.setEncoding('utf-8');
        res.on('data', function (chunk) {
            buffer += chunk;
        });
        res.on('end', function () {
            callback({ buffer: buffer, fingerprint: serverFingerprint });
        });
    });

    return req;
}

function serverSend(name, data, callback) {
    var req = bufferedRequest({ path: '/store/' + name, method: 'PUT' }, callback);
    req.write(data);
    req.end();
}

function serverList(callback) {
    bufferedRequest({ path: '/store' }, function (result) {
        if (result.buffer.length === 0) {
            con.fatal('Empty response from server - are you registered?');
        } else {
            callback(JSON.parse(result.buffer));
        }
    }).end();
}

function serverFetch(fname, callback) {
    con.debug('Get ' + fname);
    bufferedRequest({ path: '/store/' + fname }, function (result) {
        if (result.buffer.length === 0) {
            con.fatal('Empty response from server - are you registered?');
        } else {
            callback(result.buffer);
        }
    }).end();
}

// Register using a token, get certificate and key

function serverRegister(token, callback) {
    bufferedRequest({ path: '/register/' + token }, function (result) {
        if (result.buffer.length === 0) {
            con.fatal('Empty response from server - verify that the token is correct and not already used.');
        } else {
            var obj = JSON.parse(result.buffer);
            obj.fingerprint = result.fingerprint;
            callback(obj);
        }
    }).end();
}

// Get a new token that can be used to register another host

function serverToken(callback) {
    bufferedRequest({ path: '/newtoken', method: 'POST' }, function (result) {
        if (result.buffer.length === 0) {
            con.fatal('Empty response from server - are you registered?');
        } else {
            callback(JSON.parse(result.buffer));
        }
    }).end();
}

function serverNewUser(name, callback) {
    bufferedRequest({ path: '/users/' + name, method: 'POST' }, function (result) {
        if (result.buffer.length === 0) {
            con.fatal('Empty response from server - are you a registered admin?');
        } else {
            con.debug(result);
            callback(JSON.parse(result.buffer));
        }
    }).end();
}

function serverDelUser(name, callback) {
    bufferedRequest({ path: '/users/' + name, method: 'DELETE' }, callback).end();
}

function serverSaveBinary(path, local, callback) {
    unbufferedRequest({ path: path }, function (res) {
        var stream = fs.createWriteStream(local);
        res.on('data', function (chunk) {
            stream.write(chunk);
        });
        res.on('end', function () {
            stream.end();
            process.nextTick(callback);
        });
    }).end();
}

