"use strict";

var _ = require('underscore')._;
var async = require('async');
var debuggable = require('debuggable');
var express = require('express');
var fs = require('fs');
var https = require('https');
var path = require('path');
var spawn = require('child_process').spawn;
var uuid = require('node-uuid');
var mkdirp = require('mkdirp');

if (!fs.existsSync) {
    fs.existsSync = path.existsSync;
}

var Audit = require('./audit');
var con = require('./console');
var hash = require('./hash');
var UserStore = require('./users');

var minVersionExp = /^3\./;

var audit;
var users;
var caCertFile;
var serverCertFile;
var serverKeyFile;
var moleDir, dataDir, atticDir, certDir, extraDir, scriptDir;
var pkg;

exports = module.exports = server;
debuggable(server);
   
function server(opts, state) {
    audit = Audit({ auditFile: path.join(opts.store, 'audit.log') });

    verifyDependencies();

    pkg = state.pkg;
    audit.info('initializing', pkg);

    moleDir = path.resolve(opts.store);
    mkdirp.sync(moleDir);

    dataDir = path.join(moleDir, 'data');
    server.dlog('Creating data directory ' + dataDir);
    mkdirp.sync(dataDir);

    atticDir = path.join(moleDir, 'attic');
    server.dlog('Creating attic directory ' + atticDir);
    mkdirp.sync(atticDir);

    certDir = path.join(moleDir, 'crt');
    server.dlog('Creating cert directory ' + certDir);
    mkdirp.sync(certDir);

    extraDir = path.join(moleDir, 'extra');
    server.dlog('Creating extra directory ' + extraDir);
    mkdirp.sync(extraDir);

    scriptDir = path.join(__dirname, '..', 'script');

    var userFile = path.join(moleDir, 'users.json');
    server.dlog('Using users file', userFile);
    users = new UserStore(userFile);
    audit.info('users', { userFile: userFile, users: users });

    caCertFile = path.join(moleDir, 'ca-cert.pem');
    serverCertFile = path.join(certDir, 'server-cert.pem');
    serverKeyFile = path.join(certDir, 'server-key.pem');

    server.dlog('Chdir to ' + moleDir);
    process.chdir(moleDir);

    server.dlog('Checking for existing certificates');
    if (!fs.existsSync(caCertFile)) {
        con.info('Generating new CA certificate');
        server.dlog(path.join(scriptDir, 'gen-ca.exp'));
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
        server.dlog('Checking for ' + dep);
        if (!inpathSync(dep)) {
            con.fatal('Missing required dependency "' + dep + '". Please install it and retry.');
        }
    });
}

function auditReq(req) {
    return {
        ip: req.connection.socket.remoteAddress,
        port: req.connection.socket.remotePort,
        method: req.method,
        url: req.url,
        params: req.params,
        headers: req.headers
    };
}

function audited(func) {
    return function (req, res) {
        var ai = auditReq(req);
        console.log(JSON.stringify(ai));
        audit.info('request', {req:  ai});

        if (!req.headers['x-mole-version']
            || !req.headers['x-mole-version'].match(minVersionExp)) {
            res.send("Client version unacceptable", 530);
            res.end();
        } else {
            func(req, res);
        }
    };
}

function startApp(opts) {
    server.dlog('Create HTTP server');

    var app = express();
    var srv = https.createServer({
        ca: [ fs.readFileSync(caCertFile) ],
        key: fs.readFileSync(serverKeyFile),
        cert: fs.readFileSync(serverCertFile),
        requestCert: true,
    }, app);

    app.get(/\/register\/([0-9a-f\-]+)$/, audited(register));
    app.post('/newtoken', audited(newtoken));
    app.post(/\/users\/([a-z0-9_\-]+)$/, audited(newuser));
    app.del(/\/users\/([a-z0-9_\-]+)$/, audited(deluser));
    app.get('/store', audited(listfiles));
    app.put(/\/store\/([0-9a-z_.\-]+)$/, audited(putfile));
    app.get(/\/store\/([0-9a-z_.\-]+)$/, audited(getfile));
    app.get(/\/extra\/([0-9a-z_.\-]+)$/, audited(getExtraFile));
    app.get(/\/pkg/, audited(getPkg));

    srv.listen(opts.port);

    con.info('Server listening on port ' + opts.port);
    audit.info('listening', { port: opts.port });
}

function createUserCert(name, callback) {
    audit.info('createUserCert', { name: name });

    var openssl = spawn(path.join(scriptDir, 'gen-user.exp'), [ name ]);
    var fingerprint;

    function recv(data) {
        var s = data.toString('utf-8').trim();
        if (s.match(/^[0-9A-F:]+$/)) {
            fingerprint = s;
        } else if (s.length > 0) {
            audit.warning('createUserCert', { name: name, data: s });
            con.warning(s);
        }
    }

    openssl.stdout.on('data', recv);
    openssl.stderr.on('data', recv);

    openssl.on('exit', function (code) {
        audit.info('createUserCert', { name: name, fingerprint: fingerprint });
        callback(fingerprint);
    });
}

function authenticate(req) {
    if (!req.client.authorized) {
        audit.warning('unauthorized', { req: auditReq(req) });
        server.dlog('Client not authorized');
        return null;
    }

    var cert = req.connection.getPeerCertificate();
    var username = cert.subject.CN;

    var user = users.get(username);
    if (!user) {
        audit.warning('unknown username', { req: auditReq(req), username: username });
        con.warning('Certificate claimed username "' + username + '" which does not exist');
        return null;
    }

    if (user.fingerprint !== cert.fingerprint) {
        audit.warning('mismatching fingerprint', { req: auditReq(req), username: username, correct: user.fingerprint, presented: cert.fingerprint });
        con.warning('Certificate presented for "' + username + '" does not match stored fingerprint');
        return null;
    }

    audit.info('authenticated', { req: auditReq(req), username: username, fingerprint: cert.fingerprint });
    server.dlog('Certificate authentication for ' + username + ' succeeded');
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

    server.dlog('GET /register/' + token);
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

                audit.info('register successful', { req: auditReq(req), name: user.name });
                res.json({ cert: cert, key: key });
            }
        }
    });

    if (!found) {
        audit.warning('register unsuccessful', { req: auditReq(req) });
        res.send(404);
        res.end();
    }
}

function newtoken(req, res){
    server.dlog('POST /newtoken');
    var user = authenticate(req);
    if (user) {
        user.token = uuid.v4();
        users.save();
        res.json({ token: user.token });
        audit.info('new token', { req: auditReq(req), token: user.token });
    } else {
        res.send(403);
        res.end();
    }
}

// Create a new user (or reset certificate and token for an existing one).

function newuser(req, res){
    var username = req.params[0];
    server.dlog('POST /users/' + username);
    var user = authenticate(req, users);
    if (users.all().length === 0 || user && user.admin) {
        var newUserAdmin = users.all().length === 0;
        createUser(username, newUserAdmin, function (u) {
            audit.info('new user', { req: auditReq(req), username: username, admin: newUserAdmin });
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
    server.dlog('DELETE /users/' + username);
    var user = authenticate(req, users);
    if (user && user.admin) {
        if (users.get(username)) {
            users.del(username);
            users.save();
            audit.info('delete user', { req: auditReq(req), username: username });
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
    server.dlog('GET /store');
    function stat(fname, callback) {
        var fp = path.join(dataDir, fname);
        fs.stat(fp, function (err, res) {
            if (err) {
                return callback(err);
            }

            hash(fp, function(err, sha1) {
                if (err) {
                    return callback(err);
                }

                callback(null, { name: fname, mtime: res.mtime.getTime(), sha1: sha1 });
            });
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
    server.dlog('PUT /store/' + file);
    var user = authenticate(req);
    if (user) {
        var buffer = '';
        req.setEncoding('utf-8');
        req.on('data', function (chunk) {
            buffer += chunk;
        });
        req.on('end', function () {
            var fp = path.join(dataDir, file);
            if (fs.existsSync(fp)) {
                // FIXME: Only do this if the hash is different
                var bk = path.join(atticDir, file + '.' + Date.now());
                fs.renameSync(fp, bk);
                audit.info('moved', { from: fp, to: bk, req: auditReq(req) });
            }
            fs.writeFile(fp, buffer, function () {
                res.json({ status: 'ok', length: buffer.length });
                audit.info('saved', { file: fp, req: auditReq(req) });
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
    server.dlog('GET /store/' + file);
    var user = authenticate(req);
    if (user) {
        audit.info('sending', {file: file});
        res.sendfile(path.join(dataDir, file));
    } else {
        res.send(403);
        res.end();
    }
}

// Get a file from extra (read-only) storage.

function getExtraFile(req, res) {
    var file = req.params[0];
    server.dlog('GET /extra/' + file);
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
    server.dlog('GET /pkg');
    var user = authenticate(req);
    if (user) {
        res.json(pkg);
    } else {
        res.send(403);
        res.end();
    }
}

