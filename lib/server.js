#!/usr/bin/env node

var _ = require('underscore');
var https = require('https');

// FIXME: We should not use the con functions
var con = require('./console');

var defaults = {
    host: undefined,
    port: undefined,
    key: undefined,
    cert: undefined
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
}

function init(options) {
    _.extend(defaults, options);
}

function server(options, callback) {
    _.defaults(options, defaults);
    _.defaults(options, { method: 'GET', agent: false });

    var buffer = '';
    var req = https.request(options, function (res) {
        if (res.statusCode !== 200) {
            con.fatal('Server returned code ' + res.statusCode);
        }

        res.setEncoding('utf-8');
        res.on('data', function (chunk) {
            buffer += chunk;
        });
        res.on('end', function () {
            callback(buffer);
        });
    });

    req.on('error', function (err) {
        con.fatal(err);
    });

    return req;
}

function serverSend(name, data, callback) {
    var req = server({ path: '/store/' + name, method: 'PUT' }, callback);
    req.write(data);
    req.end();
}

function serverList(callback) {
    server({ path: '/store' }, function (result) {
        if (result.length === 0) {
            con.fatal('Empty response from server - are you registered?');
        } else {
            callback(JSON.parse(result));
        }
    }).end();
}

function serverFetch(fname, callback) {
    con.debug('Get ' + fname);
    server({ path: '/store/' + fname }, function (result) {
        if (result.length === 0) {
            con.fatal('Empty response from server - are you registered?');
        } else {
            callback(result);
        }
    }).end();
}

// Register using a token, get certificate and key

function serverRegister(token, callback) {
    server({ path: '/register/' + token }, function (result) {
        if (result.length === 0) {
            con.fatal('Empty response from server - verify that the token is correct and not already used.');
        } else {
            callback(JSON.parse(result));
        }
    }).end();
}

// Get a new token that can be used to register another host

function serverToken(callback) {
    server({ path: '/newtoken', method: 'POST' }, function (result) {
        if (result.length === 0) {
            con.fatal('Empty response from server - are you registered?');
        } else {
            callback(JSON.parse(result));
        }
    }).end();
}

function serverNewUser(name, callback) {
    server({ path: '/users/' + name, method: 'POST' }, function (result) {
        if (result.length === 0) {
            con.fatal('Empty response from server - are you a registered admin?');
        } else {
            con.debug(result);
            callback(JSON.parse(result));
        }
    }).end();
}

function serverDelUser(name, callback) {
    server({ path: '/users/' + name, method: 'DELETE' }, callback).end();
}
