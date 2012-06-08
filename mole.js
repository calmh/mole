#!/usr/bin/env node

var cli = require('cli').enable('status');
var https = require('https');
var nconf = require('nconf');
var fs = require('fs');
var path = require('path');
var mkdirp = require('mkdirp');
var table = require('easy-table');

function getUserHome() {
    return process.env[(process.platform == 'win32') ? 'USERPROFILE' : 'HOME'];
}

var userHome = getUserHome();
var configFile = path.join(userHome, '.mole.json');
var certFile = path.join(userHome, '.mole.crt');
var keyFile = path.join(userHome, '.mole.key');
var recipeDir = path.join(userHome, '.mole.recipes');

nconf.file({ file: configFile });
mkdirp(recipeDir);

function server(options, callback) {
    options.port = options.port || 9443;
    options.agent = false;
    try {
        options.key = fs.readFileSync(keyFile, 'utf-8');
        options.cert = fs.readFileSync(certFile, 'utf-8');
    } catch (err) {
        // We don't have a certificate yet.
    }

    var buffer = '';
    var req = https.request(options, function (res) {
        res.setEncoding('utf-8');
        res.on('data', function (chunk) {
            buffer += chunk;
        });
        res.on('end', function () {
            callback(buffer);
        });
    });
    req.end();
}

function serverSend(name, data, callback) {
    var options = { host: nconf.get('server:hostname'), path: '/store/' + name, method: 'PUT' };
    options.port = options.port || 9443;
    options.agent = false;
    try {
        options.key = fs.readFileSync(keyFile, 'utf-8');
        options.cert = fs.readFileSync(certFile, 'utf-8');
    } catch (err) {
        // We don't have a certificate yet.
    }

    var buffer = '';
    var req = https.request(options, function (res) {
        res.setEncoding('utf-8');
        res.on('data', function (chunk) {
            buffer += chunk;
        });
        res.on('end', function () {
            callback(buffer);
        });
    });

    req.write(data);
    req.end();
}

function serverRegister(host, token, callback) {
    server({ host: host, path: '/register/' + token }, function (result) {
        if (result.length === 0) {
            cli.fatal('Empty response from server - verify that the token is correct and not already used.');
        } else {
            callback(JSON.parse(result));
        }
    });
}

function serverToken(host, callback) {
    server({ host: host, path: '/newtoken', method: 'POST' }, function (result) {
        if (result.length === 0) {
            cli.fatal('Empty response from server - are you registered?');
        } else {
            callback(JSON.parse(result));
        }
    });
}

function serverList(host, callback) {
    server({ host: host, path: '/store', method: 'GET' }, function (result) {
        if (result.length === 0) {
            cli.fatal('Empty response from server - are you registered?');
        } else {
            callback(JSON.parse(result));
        }
    });
}

function serverFetch(fname, callback) {
    server({ host: nconf.get('server:hostname'), path: '/store/' + fname, method: 'GET' }, function (result) {
        if (result.length === 0) {
            cli.fatal('Empty response from server - are you registered?');
        } else {
            var local = path.join(recipeDir, fname);
            fs.writeFileSync(local, result);
            cli.ok('Fetched ' + fname);
        }
    });
}

cli.parse({ sshdebug: [ 'v', 'Enable verbose ssh sessions' ] }, ['register', 'connect', 'list', 'token', 'sync', 'put']);

cli.main(function(args, options) {
    // External modules.

    cli.debug('Starting up');
    var _ = require('underscore')._;
    var fs = require('fs');
    var kexec = require('kexec');
    var temp = require('temp');
    var util = require('util');

    if (cli.command === 'register') {
        if (this.argc !== 2) {
            cli.info('Usage: ' + cli.app + ' register <host> <token>');
            cli.fatal('Please register a mole token.');
        } else {
            cli.info('Requesting registration from server');
            serverRegister(args[0], args[1], function (result) {
                cli.info('Received certificate and key from server');
                fs.writeFileSync(certFile, result.cert);
                fs.writeFileSync(keyFile, result.key);
                fs.chmodSync(certFile, 0600);
                nconf.set('server:hostname', args[0]);
                nconf.save();
                cli.ok('Fully registered');
            });
            return;
        }
    }

    if (!path.existsSync(certFile)) {
        cli.info('Mole needs to be registered for first use.');
        cli.info('Usage: ' + cli.app + ' register <host> <token>');
        cli.fatal('Please register a mole token.');
    }

    if (cli.command === 'token') {
        cli.info('Requesting new token from server');
        serverToken(nconf.get('server:hostname'), function (result) {
            cli.ok('New token: ' + result.token);
        });
        return;
    }

    if (cli.command === 'list') {
        fs.readdir(recipeDir, function (err, files) {
            var t = new table();
            files.forEach(function (file) {
                var descr;
                try {
                    var r = require(path.join(recipeDir, file));
                    descr = r.description;
                } catch (err) {
                    descr = '(Unparseable)';
                }
                t.cell('Recipe', file.replace(/\.js$/, ''));
                t.cell('Description', descr);
                t.newLine();
            });
            console.log(t.toString());
        });
        return;
    }

    if (cli.command === 'sync') {
        cli.info('Requesting recipe list from server');
        serverList(nconf.get('server:hostname'), function (result) {
            cli.ok('List recieved');
            result.forEach(function (res) {
                var local = path.join(recipeDir, res.name);
                if (!path.existsSync(local)) {
                    cli.info(res.name + ' -- missing');
                    serverFetch(res.name);
                } else {
                    var s = fs.statSync(local);
                    if (s.mtime.getTime() < res.mtime) {
                        cli.info(res.name + ' -- out of date');
                        serverFetch(res.name);
                    } else {
                        cli.ok(res.name + ' -- in sync');
                    }
                }
            });
        });
        return;
    }

    if (cli.command === 'put') {
        cli.info('Uploading data');
        var data = fs.readFileSync(args[0], 'utf-8');
        serverSend(args[0], data, function (result) {
        });
        return;
    }

    if (cli.command === 'connect' && !this.argc) {
        cli.fatal('Usage: ' + cli.app + ' connect <destination>');
    }

    // Internal modules.

    cli.debug('Loading modules');
    var sshConfig = require('./lib/ssh-config');
    var expectConfig = require('./lib/expect-config');
    var setupLocalIPs = require('./lib/setup-local-ips');

    // Load a configuration, generate a temporary filename for ssh config.

    cli.debug('Loading configuration');
    var config = require(path.join(recipeDir, args[0]));
    config.sshConfig = temp.path({suffix: '.sshconfig'});

    // Create and save the ssh config

    cli.debug('Creating SSH configuration');
    var defaults = ['Host *', '  UserKnownHostsFile /dev/null', '  StrictHostKeyChecking no'].join('\n') + '\n';
    var conf = defaults + sshConfig(config) + '\n';
    fs.writeFileSync(config.sshConfig, conf);
    cli.debug('SSH configuration was saved to ' + config.sshConfig);
    cli.ok('Created SSH configuration');

    // Set up local IP:s needed for forwarding and execute the expect scipt.

    cli.debug('Setting up local IP:s for forwarding');
    setupLocalIPs(config, cli, options, function (c) {
        if (!c) {
            cli.error('Failed to set up IP:s for forwarding');
            cli.info('Continuing without forwarding');
            delete config.forwards;
        }

        // Create the expect script and save it to a temp file.

        cli.debug('Creating Expect script');
        var expect = expectConfig(config, cli, options) + '\n';
        var expectFile = temp.path({suffix: '.expect'});
        fs.writeFileSync(expectFile, expect);
        cli.debug('Expect script was saved to ' + expectFile);
        cli.ok('Created Expect script');

        cli.ok('Shifting into cyberspace');
        kexec('expect ' + expectFile);
    });
});

