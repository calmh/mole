"use strict";

var EventEmitter = require('events').EventEmitter;
var _ = require('underscore');
var fs = require('fs');
var https = require('https');
var util = require('util');

// The Server class is an EventEmitter, so that it can emit 'error' events when
// a problem occurs.

function Server() {
    EventEmitter.call(this);
}

util.inherits(Server, EventEmitter);
module.exports = Server;

// These are the defaults that will be used for every server call unless
// something else is explicitly specified.  They can be changed with the
// `init()` function.

var defaults = {
    host: undefined,
    port: undefined,
    key: undefined,
    cert: undefined,
    method: 'GET',
    agent: false
};

Server.prototype.init = function (options) {
    _.extend(defaults, options);
};

// Send an unbuffered request. This basically just wraps a HTTPS call with some
// scaffolding to handle errors and check the certificate. It returns the
// request `req` which you need to `.end()` when you're happy with it and calls
// the callback with the result `res` after doing som rudimentary checks on it.

Server.prototype.unbufferedRequest = function unbufferedRequest(options, callback) {
    var self = this;

    _.defaults(options, defaults);

    var req = https.request(options, function (res) {
        var serverCert = res.connection.getPeerCertificate();
        var serverFingerprint = serverCert.fingerprint;

        if (res.statusCode !== 200) {
            self.emit('error', new Error('Server returned code ' + res.statusCode));
        } else if (options.fingerprint && options.fingerprint !== serverFingerprint) {
            self.emit('error', new Error('Presented server certificate does not match stored fingerprint'));
        } else {
            callback(res);
        }
    });

    req.on('error', function (err) {
        self.emit('error', err);
    });

    return req;
};

// Send a buffered request. This takes the `unbufferedRequest` above and
// buffers all recieved data, which is expected to be UTF-8 text. When all data
// is fetched, the callback is called with the buffered text.

Server.prototype.bufferedRequest = function bufferedRequest(options, callback) {
    var buffer = '';
    var req = this.unbufferedRequest(options, function (res) {
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
};

Server.prototype.send = function send(name, data, callback) {
    var req = this.bufferedRequest({ path: '/store/' + name, method: 'PUT' }, callback);
    req.write(data);
    req.end();
};

Server.prototype.list = function list(callback) {
    var self = this;

    this.bufferedRequest({ path: '/store' }, function (result) {
        if (result.buffer.length === 0) {
            self.emit('error', new Error('Empty response from server - are you registered?'));
        } else {
            callback(JSON.parse(result.buffer));
        }
    }).end();
};

// Register using a token, get certificate and key

Server.prototype.register = function register(token, callback) {
    var self = this;

    this.bufferedRequest({ path: '/register/' + token }, function (result) {
        if (result.buffer.length === 0) {
            self.emit('error', new Error('Empty response from server - verify that the token is correct and not already used.'));
        } else {
            var obj = JSON.parse(result.buffer);
            obj.fingerprint = result.fingerprint;
            callback(obj);
        }
    }).end();
};

// Get a new token that can be used to register another host

Server.prototype.token = function token(callback) {
    var self = this;

    this.bufferedRequest({ path: '/newtoken', method: 'POST' }, function (result) {
        if (result.buffer.length === 0) {
            self.emit('error', new Error('Empty response from server - are you registered?'));
        } else {
            callback(JSON.parse(result.buffer));
        }
    }).end();
};

Server.prototype.newUser = function newUser(name, callback) {
    var self = this;

    this.bufferedRequest({ path: '/users/' + name, method: 'POST' }, function (result) {
        if (result.buffer.length === 0) {
            self.emit('error', new Error('Empty response from server - are you a registered admin?'));
        } else {
            callback(JSON.parse(result.buffer));
        }
    }).end();
};

Server.prototype.delUser = function delUser(name, callback) {
    this.bufferedRequest({ path: '/users/' + name, method: 'DELETE' }, callback).end();
};

Server.prototype.saveBin = function saveBin(path, local, callback) {
    // Add /store to unqualified paths, to be compatible with the old 'fetch' method.
    if (path.indexOf('/') < 0) {
        path = '/store/' + path;
    }

    this.unbufferedRequest({ path: path }, function (res) {
        var stream = fs.createWriteStream(local);
        res.on('data', function (chunk) {
            stream.write(chunk);
        });
        res.on('end', function () {
            // We only send the callback once the local file is actually closed.
            stream.on('close', callback);
            stream.end();
        });
    }).end();
};

Server.prototype.getPkg = function getPkg(callback) {
    this.bufferedRequest({ path: '/pkg' }, function (res) {
        callback(JSON.parse(res.buffer));
    }).end();
};

