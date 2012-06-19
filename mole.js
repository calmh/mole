#!/usr/bin/env node

var _ = require('underscore');
var exec = require('child_process').exec;
var fs = require('fs');
var inireader = require('inireader');
var iso8601 = require('iso8601');
var mkdirp = require('mkdirp');
var os = require('os');
var parser = require('nomnom');
var path = require('path');
var pidof = require('pidof');
var spawn = require('child_process').spawn;
var temp = require('temp');
var util = require('util');
var vpnc = require('vpnc');

var table = require('./lib/table');
var con = require('./lib/console');
var tun = require('./lib/tunnel');
var srv = require('./lib/server');

var configDir = path.join(process.env['HOME'], '.mole');
var configFile = path.join(configDir, 'mole.ini');
var certFile = path.join(configDir, 'mole.crt');
var keyFile = path.join(configDir, 'mole.key');
var tunnelDefDir = path.join(configDir, 'tunnels');
var pkgDir = path.join(configDir, 'pkg');

mkdirp.sync(tunnelDefDir);
mkdirp.sync(pkgDir);
fs.chmodSync(configDir, 0700);

var config = new inireader.IniReader();
try {
    config.load(configFile);
} catch (err) {
    con.info('No config, using defaults.');
    config.param('server.port', 9443);
    config.write();
}

parser.script('mole');

parser.command('dig')
.help('Dig a tunnel to the destination')
.option('tunnel', { position: 1, help: 'Tunnel name or file name', required: true })
.option('host', { position: 2, help: 'Host name within tunnel definition' })
.callback(dig);

parser.command('list')
.help('List available tunnel definitions')
.callback(list);

parser.command('pull')
.help('Get tunnel definitions from the server')
.callback(pull);

parser.command('push')
.help('Send a tunnel definition to the server')
.option('file', { position: 1, help: 'File name', required: true })
.callback(push);

parser.command('register')
.help('Register with a mole server')
.option('server', { position: 1, help: 'Server name', required: true })
.option('token', { position: 2, help: 'One time registration token', required: true })
.option('port', { abbr: 'p', metafile: 'PORT', help: 'Server port (default 9443)', default: 9443 })
.callback(register);

parser.command('gettoken')
.help('Generate a new registration token')
.callback(token);

parser.command('export')
.option('tunnel', { position: 1, help: 'Tunnel name', required: true })
.option('file', { position: 2, help: 'File name to write tunnel definition to', required: true })
.help('Export tunnel definition to a file')
.callback(exportf);

parser.command('newuser')
.option('name', { position: 1, help: 'User name', required: true })
.option('admin', { flag: true, abbr: 'a', help: 'Create an admin user' })
.help('Create a new user (requires admin privileges)')
.callback(newUser);

parser.command('deluser')
.option('name', { position: 1, help: 'User name', required: true })
.help('Delete a user (requires admin privileges)')
.callback(delUser);

parser.command('install')
.option('package', { position: 1, help: 'Package name', required: true })
.callback(install);

parser.option('debug', { abbr: 'd', flag: true, help: 'Display debug output' });

parser.option('help', { abbr: 'h', flag: true, help: 'Display command help' });

parser.nocommand().callback(function () {
    console.log(parser.getUsage());
    process.exit(0);
})
.help([
      'Examples:',
      '',
      'Register with server "mole.example.com" and a token:',
      '  mole register mole.example.com 80721953-b4f2-450e-aaf4-a1c0c7599ec2',
      '',
      'List available tunnels:',
      '  mole list',
      '',
      'Dig a tunnel to "operator3":',
      '  mole dig operator3',
      '',
      'Fetch new and updated tunnel specifications from the server:',
      '  mole pull'
].join('\n'));

parser.parse();

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

function init(opts) {
    if (opts.debug) {
        con.enableDebug();
    }

    srv.init({
        host: config.param('server.host'),
        port: config.param('server.port') || 9443,
        fingerprint: config.param('server.fingerprint')
    });

    readCert();
};

function register(opts) {
    init(opts);

    con.debug('Requesting registration from server ' + opts.server + ':' + opts.port);
    config.param('server.host', opts.server);
    config.param('server.port', opts.port);
    srv.init({ host: opts.server, port: opts.port });
    srv.register(opts.token, function (result) {
        con.debug('Received certificate and key from server');
        fs.writeFileSync(certFile, result.cert);
        fs.writeFileSync(keyFile, result.key);
        config.param('server.fingerprint', result.fingerprint);
        config.write();
        con.ok('Registered');
        readCert();
        pull(opts);
    });
}

function token(opts) {
    init(opts);

    con.debug('Requesting new token from server');
    con.info('A token can be used only once');
    con.info('Only the most recently generated token is valid');
    srv.token(function (result) {
        con.ok(result.token);
    });
}

function list(opts) {
    init(opts);

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

function pull(opts) {
    init(opts);

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

function push(opts) {
    init(opts);

    con.debug('Reading ' + opts.file);
    try {
        var data = fs.readFileSync(opts.file, 'utf-8');
        con.debug('Got ' + data.length + ' bytes');
    } catch (err) {
        con.fatal(err);
    }

    con.debug('Testing ' + opts.file);
    try {
        tun.load(opts.file);
        con.debug('It passed validation');
    } catch (err) {
        con.fatal(err);
    }

    srv.send(path.basename(opts.file), data, function (result) {
        con.ok('Sent ' + data.length + ' bytes');
    });
}

function newUser(opts) {
    init(opts);

    con.debug('Requesting user ' + opts.name);
    srv.newUser(opts.name, function (result) {
        con.ok(result.token);
    });
}

function delUser(opts) {
    init(opts);

    con.debug('Deleting user ' + opts.name);
    srv.delUser(opts.name, function (result) {
        con.ok('deleted');
    });
}

function dig(opts) {
    init(opts);

    exec('expect -v', function (error, stdout, stderr) {
        if (error) {
            con.error(error.toString().trim());
            con.error('Verify that "expect" is installed and available in the path.');
        } else {
            con.debug(stdout.trim());
            digReal(opts.tunnel, opts.host, opts.debug);
        }
    });
}

function digReal(tunnel, host, debug) {
    var config;
    var sshConfig = require('./lib/ssh-config');

    // Load a configuration, generate a temporary filename for ssh config.

    con.debug('Loading tunnel');
    try {
        config = tun.load(tunnel, tunnelDefDir);
    } catch (err) {
        con.fatal(err);
    }
    config.sshConfig = temp.path({suffix: '.ssh.conf'});

    // Create and save the ssh config

    con.debug('Creating ssh configuration');
    var defaults = [
        'Host *',
        '  UserKnownHostsFile /dev/null',
        '  StrictHostKeyChecking no',
        '  IdentitiesOnly yes'
    ].join('\n') + '\n';
    var conf = defaults + sshConfig(config);
    fs.writeFileSync(config.sshConfig, conf);
    con.debug(config.sshConfig);

    if (config.vpnc) {
        vpnc.available(function (err, result) {
            if (err) {
                con.error(err);
                con.error('vpnc unavailable; try "mole install vpnc"');
                con.fatal('Not continuing without vpnc');
            } else {
                con.debug('Using ' +  result.version);

                pidof('vpnc', function (err, pid) {
                    if (err) {
                        con.error(err);
                        con.fatal('could not check if vpnc was running');
                    } else if (pid) {
                        con.fatal('vpnc already running, will not start another instance');
                    }

                    con.info('Connecting VPN; you might be asked for your local (sudo) password now');
                    console.log(config);
                    vpnc.connect(config.vpnc, config.vpnRoutes, function (err, code) {
                        if (err) {
                            con.fatal(err);
                        } else if (code !== 0) {
                            con.fatal('vpnc returned an error - investigate and act on it, nothing more I can do :(');
                        }
                        con.info('VPN connected. Should the login sequence fail, you can disconnect the VPN');
                        con.info('manually by running "sudo ' + result.vpncDisconnect + '"');
                        launchExpect(config, debug, host);
                    });
                });
            }
        });
    } else {
        launchExpect(config, debug, host);
    }
};

function launchExpect(config, debug, host) {
    var setupLocalIPs = require('./lib/setup-local-ips');
    var expectConfig = require('./lib/expect-config');

    con.debug('Creating expect script');
    try {
        var expect = expectConfig(config, debug, host);
        var expectFile = temp.path({suffix: '.expect'});
        fs.writeFileSync(expectFile, expect);
        con.debug(expectFile);
    } catch (err) {
        con.fatal(err);
    }

    // Set up local IP:s needed for forwarding and execute the expect scipt.

    con.debug('Setting up local IP:s for forwarding');
    setupLocalIPs(config, function (c) {
        if (!c) {
            con.warning('Failed to set up IP:s for forwarding. Continuing without forwarding.');
            delete config.forwards;
        }

        // Create the expect script and save it to a temp file.

        con.info('Hang on, digging the tunnel');
        spawn('expect', [ expectFile ], { customFds: [ 0, 1, 2 ] })
        .on('exit', function (code) {
            fs.unlinkSync(expectFile);
            fs.unlinkSync(config.sshConfig);
            // FIXME: Unlink ssh keys

            if (config.vpnc) {
                con.debug('Disconnecting VPN');
                vpnc.disconnect(function (err, status) {
                    if (err) {
                        con.fatal(err);
                    }
                    con.info('VPN disconnected');
                });
            }

            if (code === 0) {
                con.ok('Great success');
            } else {
                con.error('Unsuccessful');
                con.info('Debug tunnel definition by "mole dig -d <tunnel>" or talk to the author:');
                con.info(config.author);
            }
        });
    });
}

function exportf(opts) {
    var config;

    init(opts);

    try {
        con.debug('Loading tunnel');
        config = tun.load(opts.tunnel, tunnelDefDir);
    } catch (err) {
        con.fatal(err);
    }

    con.debug('Saving to INI format');
    tun.save(config, opts.file);

    con.ok(opts.file);
};

function install(opts) {
    init(opts);

    var file = [ opts.package, os.platform(), os.arch(), os.release() ].join('-') + '.tar.gz';
    var local = path.join(pkgDir, file);
    var tmp = temp.path();
    mkdirp(tmp);

    con.info('Fetching ' + file);
    srv.saveBin('/extra/' + file, local, function () {
        con.info('Unpacking ' + file);
        exec('cd ' + tmp + ' && tar zxf ' + local, function (err, stdout, stderr) {
            con.debug('Extracted in ' + tmp);
            con.info('Running installation, you might now be asked for your local (sudo) password.');
            var inst = spawn('sudo', [ path.join(tmp, 'install.sh'), tmp ], { customFds: [ 0, 1, 2 ] });
            inst.on('exit', function (code) {
                if (code === 0) {
                    con.info('Installation complete');
                } else {
                    con.info('Installation failed. Sorry.');
                }
            });
        });
    }, local);
};

