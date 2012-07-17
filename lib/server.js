"use strict";

var _ = require('underscore')._;
var async = require('async');
var express = require('express')
var fs = require('fs');
var https = require('https');
var path = require('path');
var spawn = require('child_process').spawn;
var uuid = require('node-uuid');
var mkdirp = require('mkdirp');
if (!fs.existsSync) {
    fs.existsSync = path.existsSync;
}

var userStore = require('./users');
var con = require('./console');

var users;
var caCertFile;
var serverCertFile;
var serverKeyFile;
var moleDir, dataDir, certDir, extraDir, scriptDir;
var pkg;

module.exports = function init(opts, state) {
    con.enableTimestamps();
    if (opts.debug) {
        con.enableDebug();
    }

    verifyDependencies();

    pkg = state.pkg;

    moleDir = path.resolve(opts.store);

    dataDir = path.join(moleDir, 'data');
    con.debug('Creating data directory ' + dataDir);
    mkdirp.sync(dataDir);

    certDir = path.join(moleDir, 'crt');
    con.debug('Creating cert directory ' + certDir);
    mkdirp.sync(certDir);

    extraDir = path.join(moleDir, 'extra');
    con.debug('Creating extra directory ' + extraDir);
    mkdirp.sync(extraDir);

    scriptDir = path.join(__dirname, '..', 'script');

    var userFile = path.join(moleDir, 'users.json');
    con.debug('Using users file', userFile);
    users = new userStore(userFile);

    caCertFile = path.join(moleDir, 'ca-cert.pem');
    serverCertFile = path.join(certDir, 'server-cert.pem');
    serverKeyFile = path.join(certDir, 'server-key.pem');

    con.debug('Chdir to ' + moleDir);
    process.chdir(moleDir);

    con.debug('Checking for existing certificates');
    if (!fs.existsSync(caCertFile)) {
        con.info('Generating new CA certificate');
        con.debug(path.join(scriptDir, 'gen-ca.exp'));
        spawn(path.join(scriptDir, 'gen-ca.exp')).on('exit', function (code) {
            if (code !== 0) {
                con.fatal('Error ' + code);
            }
            con.ok('Generated CA certificate');

            con.info('Generating new server certificate');
            spawn(path.join(scriptDir, 'gen-user.exp'), [ 'server' ]).on('exit', function (code) {
                if (code !== 0) {
                    con.fatal('Error ' + code);
                }
                con.ok('Generated server certificate');

                startApp(opts);
            });
        });
    } else {
        startApp(opts);
    }
}

function verifyDependencies() {
    var inpathSync = require('inpath').sync;
    [ 'expect', 'openssl' ].forEach(function (dep) {
        con.debug('Checking for ' + dep);
        if (!inpathSync(dep)) {
            con.fatal('Missing required dependency "' + dep + '". Please install it and retry.');
        }
    });
}

function startApp(opts) {
    con.debug('Create HTTP server');
    var app = express.createServer({
        ca: [ fs.readFileSync(caCertFile) ],
        key: fs.readFileSync(serverKeyFile),
        cert: fs.readFileSync(serverCertFile),
        requestCert: true,
    });

    app.get(/\/register\/([0-9a-f-]+)$/, register);
    app.post('/newtoken', newtoken);
    app.post(/\/users\/([a-z0-9_-]+)$/, newuser);
    app.del(/\/users\/([a-z0-9_-]+)$/, deluser);
    app.get('/store', listfiles);
    app.put(/\/store\/([0-9a-z_.-]+)$/, putfile);
    app.get(/\/store\/([0-9a-z_.-]+)$/, getfile);
    app.get(/\/extra\/([0-9a-z_.-]+)$/, getExtraFile);
    app.get(/\/pkg/, getPkg);
    app.listen(opts.port);
    con.info('Server listening on port ' + opts.port);
}

function createUserCert(name, callback) {
    var openssl = spawn(path.join(scriptDir, 'gen-user.exp'), [ name ]);
    var fingerprint;

    function recv(data) {
        var s = data.toString('utf-8').trim();
        if (s.match(/^[0-9A-F:]+$/)) {
            fingerprint = s;
        } else if (s.length > 0) {
            con.warning(s);
        }
    }

    openssl.stdout.on('data', recv);
    openssl.stderr.on('data', recv);

    openssl.on('exit', function (code) {
        callback(fingerprint);
    });
}

function authenticate(req) {
    if (!req.client.authorized) {
        con.debug('Client not authorized');
        return null;
    }

    var cert = req.connection.getPeerCertificate();
    var username = cert.subject.CN;

    var user = users.get(username);
    if (!user) {
        con.warning('Certificate claimed username "' + username + '" which does not exist');
        return null;
    }

    if (user.fingerprint !== cert.fingerprint) {
        con.warning('Certificate presented for "' + username + '" does not match stored fingerprint');
        return null;
    }

    con.debug('Certificate authentication for ' + username + ' succeeded');
    return user;
}

function createUser(name, admin, callback) {
    var user = { created: Date.now(), token: uuid.v4() };

    createUserCert(name, function (fingerprint) {
        user.fingerprint = fingerprint;
        user.admin = !!admin;
        users.set(name, user);
        callback(user);
    });
}

function register(req, res){
    var token = req.params[0];
    var found = false;

    con.debug('GET /register/' + token);
    _.each(users.all(), function (user) {
        if (!found) {
            var name = user.name;
            var ud = user.data;
            if (ud.token === token) {
                found = true;
                var cert = fs.readFileSync('crt/' + name + '-cert.pem', 'utf-8');
                var key = fs.readFileSync('crt/' + name + '-key.pem', 'utf-8');
                delete ud.token;
                ud.registered = Date.now();
                users.save();

                res.json({ cert: cert, key: key });
            }
        }
    });

    if (!found) {
        res.send(404);
        res.end();
    }
}

function newtoken(req, res){
    con.debug('POST /newtoken');
    var user = authenticate(req);
    if (user) {
        user.token = uuid.v4();
        users.save();
        res.json({ token: user.token });
    } else {
        res.send(403);
        res.end();
    };
}

// Create a new user (or reset certificate and token for an existing one).

function newuser(req, res){
    var username = req.params[0];
    con.debug('POST /users/' + username);
    var user = authenticate(req, users);
    if (users.all().length === 0 || user && user.admin) {
        var newUserAdmin = users.all().length == 0;
        createUser(username, newUserAdmin, function (u) {
            res.send(JSON.stringify(u));
        });
    } else {
        res.send(403);
        res.end();
    }
}

// Delete a user

function deluser(req, res){
    var username = req.params[0];
    con.debug('DELETE /users/' + username);
    var user = authenticate(req, users);
    if (user && user.admin) {
        if (users.get(username)) {
            users.del(username);
            users.save();
        } else {
            res.send(404);
        }
    } else {
        res.send(403);
    }
    res.end();
}

// List the files in storage.

function listfiles(req, res) {
    con.debug('GET /store');
    function stat(fname, callback) {
        fs.stat(path.join(dataDir, fname), function (err, res) {
            if (err) {
                callback(err);
            } else {
                callback(null, { name: fname, mtime: res.mtime.getTime() });
            }
        });
    }

    var user = authenticate(req);
    if (user) {
        fs.readdir(dataDir, function (err, files) {
            async.map(files, stat, function (err, files) {
                res.json(files);
            });
        });
    } else {
        res.send(403);
        res.end();
    }
}

// Add a file to storage.

function putfile(req, res) {
    var file = req.params[0];
    con.debug('PUT /store/' + file);
    var user = authenticate(req);
    if (user) {
        var buffer = '';
        req.setEncoding('utf-8');
        req.on('data', function (chunk) {
            buffer += chunk;
        });
        req.on('end', function () {
            fs.writeFile(path.join(dataDir, file), buffer, function () {
                res.json({ status: 'ok', length: buffer.length });
            });
        });
    } else {
        res.send(403);
        res.end();
    }
}

// Get a file from storage.

function getfile(req, res) {
    var file = req.params[0];
    con.debug('GET /store/' + file);
    var user = authenticate(req);
    if (user) {
        res.sendfile(path.join(dataDir, file));
    } else {
        res.send(403);
        res.end();
    }
}

// Get a file from extra (read-only) storage.

function getExtraFile(req, res) {
    var file = req.params[0];
    con.debug('GET /extra/' + file);
    var user = authenticate(req);
    if (user) {
        res.sendfile(path.join(extraDir, file));
    } else {
        res.send(403);
        res.end();
    }
}

// Get a file from extra (read-only) storage.

function getPkg(req, res) {
    con.debug('GET /pkg');
    var user = authenticate(req);
    if (user) {
        res.json(pkg);
    } else {
        res.send(403);
        res.end();
    }
}

