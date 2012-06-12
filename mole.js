#!/usr/bin/env node

var _ = require('underscore');
var commander = require('commander');
var exec = require('child_process').exec;
var fs = require('fs');
var inireader = require('inireader');
var iso8601 = require('iso8601');
var mkdirp = require('mkdirp');
var path = require('path');
var spawn = require('child_process').spawn;
var temp = require('temp');
var util = require('util');

var table = require('./lib/table');
var con = require('./lib/console');
var tun = require('./lib/tunnel');
var srv = require('./lib/server');

var configDir = path.join(process.env['HOME'], '.mole');
var configFile = path.join(configDir, 'mole.ini');
var certFile = path.join(configDir, 'mole.crt');
var keyFile = path.join(configDir, 'mole.key');
var tunnelDefDir = path.join(configDir, 'tunnels');

mkdirp.sync(tunnelDefDir);

var config = new inireader.IniReader();
try {
    config.load(configFile);
} catch (err) {
    con.info('No config, using defaults.');
    config.param('server.port', 9443);
    config.write();
}

commander
.command('dig <destination> [host]')
.description('dig a tunnel to the destination')
.action(dig);

commander
.command('list')
.description('list available tunnel definitions')
.action(list);

commander
.command('pull')
.description('get tunnel definitions from the server')
.action(pull);

commander
.command('push <file>')
.description('send a tunnel definition to the server')
.action(push);

commander
.command('register <server> <token>')
.description('register with a mole server')
.option('-p, --port', 'server port', config.param('server.port'))
.action(register);

commander
.command('gettoken')
.description('generate a new registration token')
.action(token);

commander
.command('newuser <username>')
.description('create a new user')
.option('-a, --admin', 'create an admin user')
.action(newUser);

commander
.command('export <tunnel> <outfile>')
.description('export tunnel definition to a file')
.action(exportf);

commander
.command('view <tunnel>')
.description('show tunnel definition')
.action(view);

commander
.option('-d, --debug', 'display debug information')
.on('--help', function () {
    console.log('  Examples:');
    console.log();
    console.log('    Register with server "mole.example.com" and a token:');
    console.log('      ' + commander.name + ' register mole.example.com 80721953-b4f2-450e-aaf4-a1c0c7599ec2');
    console.log();
    console.log('    List available tunnels:');
    console.log('      ' + commander.name + ' list');
    console.log();
    console.log('    Dig a tunnel to "operator3":');
    console.log('      ' + commander.name + ' dig operator3');
    console.log();
    console.log('    Fetch new and updated tunnel specifications from the server:');
    console.log('      ' + commander.name + ' pull');
    console.log();
});

commander.parse(process.argv);

// There should be a 'command' (typeof 'object') among the arguments
if (!_.any(commander.args, function (a) { return typeof a === 'object' })) {
    process.stdout.write(commander.helpInformation());
    commander.emit('--help');
    process.exit(0);
}

function readCert() {
    try {
        var key, cert;
        con.debug('Trying to load ' + keyFile);
        key = fs.readFileSync(keyFile, 'utf-8');
        con.debug('Trying to load ' + certFile);
        cert = fs.readFileSync(certFile, 'utf-8');
        srv.init({ key: key, cert: cert });
    } catch (err) {
        con.debug('No certificate loaded');
        // We don't have a certificate yet.
    }
}

function init() {
    if (commander.debug) {
        con.enableDebug();
    }

    srv.init({ host: config.param('server.host'), port: config.param('server.port') || 9443 });

    readCert();
};

function register(host, token) {
    init();

    con.debug('Requesting registration from server ' + host);
    config.param('server.host', host);
    srv.init({ host: host });
    srv.register(token, function (result) {
        con.debug('Received certificate and key from server');
        fs.writeFileSync(certFile, result.cert);
        fs.chmodSync(certFile, 0600);
        fs.writeFileSync(keyFile, result.key);
        fs.chmodSync(keyFile, 0600);
        config.write();
        con.ok('Registered');
        readCert();
        pull();
    });
}

function token() {
    init();

    con.debug('Requesting new token from server');
    con.info('A token can be used only once');
    con.info('Only the most recently generated token is valid');
    srv.token(function (result) {
        con.ok(result.token);
    });
}

function list() {
    init();

    con.debug('listing files in ' + tunnelDefDir);
    fs.readdir(tunnelDefDir, function (err, files) {
        con.debug('Got ' + files.length + ' files');

        var rows = [];
        files.sort().forEach(function (file) {
            var tname = tun.name(file);
            try {
                var r = tun.load(path.join(tunnelDefDir, file));
                var descr = r.description;
                var hosts = _.keys(r.hosts).sort().join(', ');
                var mtime = r.stat.mtime;
                rows.push([ tname, descr, hosts, iso8601.fromDate(mtime).slice(0, 10) ]);
            } catch (err) {
                rows.push([ tname, '--Corrupt--', '--Corrupt--', '--Corrupt--' ]);
            }
        });

        table([ 'Tunnel', 'Description', 'Hosts', 'Modified' ], rows);
    });
}

function pull() {
    init();

    con.debug('Requesting tunnel list from server');
    srv.list(function (result) {
        con.debug('Got ' + result.length + ' entries');
        con.debug(util.inspect(result));
        var inProgress = 0;

        function done() {
            inProgress -= 1;
            if (inProgress === 0) {
                con.ok(result.length + ' tunnel definitions in sync');
                inProgress = -1;
            }
        };

        _.sortBy(result, 'name').forEach(function (res) {
            var local = path.join(tunnelDefDir, res.name);
            inProgress += 1;

            var fetch = false;
            if (!path.existsSync(local)) {
                fetch = true;
            } else {
                var s = fs.statSync(local);
                if (s.mtime.getTime() !== parseInt(res.mtime, 10)) {
                    fetch = true;
                }
            }

            if (fetch) {
                srv.fetch(res.name, function (result) {
                    var mtime = Math.floor(res.mtime / 1000);
                    fs.writeFileSync(local, result);
                    fs.utimesSync(local, mtime, mtime);
                    con.ok('Pulled ' + tun.name(res.name));
                    done();
                });
            } else {
                process.nextTick(done);
            }
        });
    });
}

function push(file) {
    init();

    con.debug('Reading ' + file);
    try {
        var data = fs.readFileSync(file, 'utf-8');
        con.debug('Got ' + data.length + ' bytes');
    } catch (err) {
        con.fatal(err);
    }

    con.debug('Testing ' + file);
    try {
        tun.load(file);
        con.debug('It passed validation');
    } catch (err) {
        con.fatal(err);
    }

    srv.send(path.basename(file), data, function (result) {
        con.ok('Sent ' + data.length + ' bytes');
    });
}

function newUser(name) {
    init();

    con.debug('Requesting user ' + name);
    srv.newUser(name, function (result) {
        con.ok(result.token);
    });
}

function dig(tunnel, host) {
    init();

    exec('expect -v', function (error, stdout, stderr) {
        if (error) {
            con.error(error.toString().trim());
            con.error('Verify that "expect" is installed and available in the path.');
        } else {
            con.debug(stdout.trim());
            digReal(tunnel, host);
        }
    });
}

function digReal(tunnel, host) {
    var config;
    var sshConfig = require('./lib/ssh-config');
    var expectConfig = require('./lib/expect-config');
    var setupLocalIPs = require('./lib/setup-local-ips');

    // Load a configuration, generate a temporary filename for ssh config.

    con.debug('Loading tunnel');
    try {
        config = tun.load(tunnel, tunnelDefDir);
    } catch (err) {
        con.fatal(err);
    }
    config.sshConfig = temp.path({suffix: '.sshconfig'});

    // Create and save the ssh config

    con.debug('Creating ssh configuration');
    var defaults = ['Host *',
        '  UserKnownHostsFile /dev/null',
        '  StrictHostKeyChecking no',
        '  IdentitiesOnly yes'].join('\n') + '\n';
    var conf = defaults + sshConfig(config) + '\n';
    fs.writeFileSync(config.sshConfig, conf);
    con.debug(config.sshConfig);

    // Set up local IP:s needed for forwarding and execute the expect scipt.

    con.debug('Setting up local IP:s for forwarding');
    setupLocalIPs(config, function (c) {
        if (!c) {
            con.warning('Failed to set up IP:s for forwarding. Continuing without forwarding');
            delete config.forwards;
        }

        // Create the expect script and save it to a temp file.

        con.debug('Creating expect script');
        try {
            var expect = expectConfig(config, commander.debug, host) + '\n';
            var expectFile = temp.path({suffix: '.expect'});
            fs.writeFileSync(expectFile, expect);
            con.debug(expectFile);
        } catch (err) {
            con.fatal(err);
        }

        con.info('Hang on, digging the tunnel');
        spawn('expect', [ expectFile ], { customFds: [ 0, 1, 2 ] })
        .on('exit', function (code) {
            fs.unlinkSync(expectFile);
            fs.unlinkSync(config.sshConfig);
            // FIXME: Unlink ssh keys
            if (code === 0) {
                con.ok('Great success');
            } else {
                con.error('Unsuccessful');
                con.info('Debug tunnel definition by digging with -d, or talk to the author:');
                con.info(config.author);
            }
        });
    });
};

function exportf(tunnel, outfile) {
    var config;

    init();

    try {
        con.debug('Loading tunnel');
        config = tun.load(tunnel, tunnelDefDir);
    } catch (err) {
        con.fatal(err);
    }

    con.debug('Saving to INI format');
    tun.save(config, outfile);

    con.ok(outfile);
};

function view(tunnel) {
    var config;

    init();

    try {
        con.debug('Loading tunnel');
        config = tun.load(tunnel, tunnelDefDir);
    } catch (err) {
        con.fatal(err);
    }

    con.debug('Saving to INI format');
    var path = temp.path({ suffix: '.ini' });
    tun.save(config, path);

    con.debug('Show ' + path);
    console.log(fs.readFileSync(path, 'utf-8'));

    fs.unlinkSync(path);
};

