"use strict";

var _ = require('underscore')._;
var con = require('yacon');
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
var JSONStore = require('./jsonstore');
var obfuscate = require('./obfuscate');
var tunnel = require('./tunnel');

var minVersionExp = /^3\./;

var audit;
var users;
var keys;
var caCertFile;
var serverCertFile;
var serverKeyFile;
var moleDir, dataDir, atticDir, certDir, extraDir, scriptDir;
var pkg;

exports = module.exports = server;

function server(opts, state) {
    audit = Audit({ auditFile: path.join(opts.store, 'audit.log') });

    verifyDependencies();

    pkg = state.pkg;
    audit.info('initializing', pkg);

    moleDir = path.resolve(opts.store);
    mkdirp.sync(moleDir);

    dataDir = path.join(moleDir, 'data');
    con.debug('Creating data directory ' + dataDir);
    mkdirp.sync(dataDir);

    atticDir = path.join(moleDir, 'attic');
    con.debug('Creating attic directory ' + atticDir);
    mkdirp.sync(atticDir);

    certDir = path.join(moleDir, 'crt');
    con.debug('Creating cert directory ' + certDir);
    mkdirp.sync(certDir);

    extraDir = path.join(moleDir, 'extra');
    con.debug('Creating extra directory ' + extraDir);
    mkdirp.sync(extraDir);

    scriptDir = path.join(__dirname, '..', 'script');

    var userFile = path.join(moleDir, 'users.json');
    con.debug('Using users file', userFile);
    users = new JSONStore(userFile);
    audit.info('users', { userFile: userFile, items: users });

    var keysFile = path.join(moleDir, 'keys.json');
    con.debug('Using keys file', keysFile);
    keys = new JSONStore(keysFile);
    audit.info('keys', { keysFile: keysFile, items: keys });

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

function auditReq(req) {
    return {
        ip: req.connection.socket.remoteAddress,
        port: req.connection.socket.remotePort,
        method: req.method,
        url: req.url,
        params: req.params,
        headers: req.headers,
        user: req.moleAuthenticated
    };
}

function audited(func) {
    return function (req, res) {
        var ai = auditReq(req);
        console.log(JSON.stringify(ai));
        audit.info('request', {req:  ai});
        return func(req, res);
    };
}

function authenticated(func) {
    return function (req, res) {
        res.header('x-mole-protocol', '3');

        if (!req.headers['x-mole-version']
            || !req.headers['x-mole-version'].match(minVersionExp)) {
            res.send('Client version unacceptable', 530);
            res.end();
            return;
        }

        var user = authenticate(req);
        if (user)
            res.header('x-mole-authenticated', user);
        req.moleAuthenticated = user;

        return func(req, res);
    };
}

function authenticate(req) {
    if (!req.client.authorized) {
        audit.warning('unauthorized', { req: auditReq(req) });
        con.debug('Client not authorized');
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
    con.debug('Certificate authentication for ' + username + ' succeeded');

    return username;
}

function startApp(opts) {
    con.debug('Create HTTP server');

    var app = express();
    var srv = https.createServer({
        ca: [ fs.readFileSync(caCertFile) ],
        key: fs.readFileSync(serverKeyFile),
        cert: fs.readFileSync(serverCertFile),
        requestCert: true
    }, app);

    app.get(/\/register\/([0-9a-f\-]+)$/, authenticated(audited(register)));
    app.post('/newtoken', authenticated(audited(newtoken)));
    app.post(/\/users\/([a-z0-9_\-]+)$/, authenticated(audited(newuser)));
    app.del(/\/users\/([a-z0-9_\-]+)$/, authenticated(audited(deluser)));
    app.get('/store', authenticated(audited(listfiles)));
    app.put(/\/store\/([0-9a-z_.\-]+)$/, authenticated(audited(putfile)));
    app.get(/\/store\/([0-9a-z_.\-]+)$/, authenticated(audited(getfile)));
    app.get(/\/extra\/([0-9a-z_.\-]+)$/, authenticated(audited(getExtraFile)));
    app.get(/\/pkg/, authenticated(audited(getPkg)));
    app.get(/\/key\/([0-9a-f-]+)/, authenticated(audited(getKey)));

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

    openssl.on('exit', function (/*code*/) {
        audit.info('createUserCert', { name: name, fingerprint: fingerprint });
        callback(fingerprint);
    });
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
    con.debug('POST /newtoken');
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
    if (users.all().length > 0 && (!req.moleAuthenticated || !users.get(req.moleAuthenticated).admin)) {
        res.send(403);
        res.end();
        return;
    }

    var username = req.params[0];
    con.debug('POST /users/' + username);
    var newUserAdmin = users.all().length === 0;
    createUser(username, newUserAdmin, function (u) {
        audit.info('new user', { req: auditReq(req), username: username, admin: newUserAdmin });
        res.send(JSON.stringify(u));
    });
}

// Delete a user

function deluser(req, res){
    if (!req.moleAuthenticated || !users.get(req.moleAuthenticated).admin) {
        res.send(403);
        res.end();
        return;
    }

    var username = req.params[0];
    con.debug('DELETE /users/' + username);
    if (users.get(username)) {
        users.del(username);
        users.save();
        audit.info('delete user', { req: auditReq(req), username: username });
    } else {
        res.send(404);
    }
    res.end();
}

// List the files in storage.

function listfiles(req, res) {
    if (!req.moleAuthenticated) {
        res.send(403);
        res.end();
        return;
    }

    con.debug('GET /store');
    function load(fname) {
        var fp = path.join(dataDir, fname);
        try {
            var t = tunnel.load(fp);
            return {
                name: fname.replace(/\.ini$/, ''),
                description: t.general.description,
                vpnc: !!t.vpnc,
                openconnect: !!t.openconnect,
                socks: !!(t.general.main && t.hosts[t.general.main].socks),
                hosts: Object.keys(t.hosts).sort(),
                localOnly: Object.keys(t.hosts).length === 0 && Object.keys(t.forwards).length > 0
            };
        } catch (e) {
            return { name: fname, error: e.toString() };
        }
    }

    fs.readdir(dataDir, function (err, files) {
        var tunnels = files.map(load);
        res.json(tunnels);
    });
}

// Add a file to storage.

function putfile(req, res) {
    if (!req.moleAuthenticated) {
        res.send(403);
        res.end();
        return;
    }

    var file = req.params[0];
    con.debug('PUT /store/' + file);
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

        var cfg = obfuscate.obfuscate(tunnel.parse(buffer), keys);
        tunnel.save(cfg, fp);
        res.json({ status: 'ok', length: buffer.length });
        audit.info('saved', { file: fp, req: auditReq(req) });
    });
}

// Get a file from storage.

function getfile(req, res) {
    if (!req.moleAuthenticated) {
        res.send(403);
        res.end();
        return;
    }

    var file = req.params[0];
    con.debug('GET /store/' + file);
    audit.info('sending', {file: file});
    res.sendfile(path.join(dataDir, file));
}

// Get a file from extra (read-only) storage.

function getExtraFile(req, res) {
    if (!req.moleAuthenticated) {
        res.send(403);
        res.end();
        return;
    }

    var file = req.params[0];
    con.debug('GET /extra/' + file);
    res.sendfile(path.join(extraDir, file));
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

// Get a key from the keystore

function getKey(req, res) {
    if (!req.moleAuthenticated) {
        res.send(403);
        res.end();
        return;
    }

    var key = req.params[0];
    con.debug('GET /key/' + key);
    res.json({key: keys.get(key)});
}

