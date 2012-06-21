#!/usr/bin/env node

// We now use strict for everything since Node supports it.

"use strict";

// Require a whole bunch of external libraries.

var _ = require('underscore');
var colors = require('colors');
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
var table = require('yatf');
var temp = require('temp');
var util = require('util');
var version = require('version');
var vpnc = require('vpnc');

// Figure out if we are running with a TTY on stdout or not.
// If not, we'll avoid using ANSI colors later on.

var isatty = process.stdout.isTTY;

// The existsSync function moved between Node 0.6 and 0.8.  We just use it from
// wherever we found it.
//
// You'll be seeing a lot of `*Sync` calls in here. If that disturbs you, keep
// in mind that this is a CLI program that runs once and then exits, not some
// sort of high performance IO-bound server code, mmkay?

var existsSync = fs.existsSync; // Node 0.8
if (!existsSync) {
    existsSync = path.existsSync; // Node 0.6 and prior
}

// We load our own package file to get at the version number.

var pkg = require(path.join(__dirname, 'package.json'));

// Load internal modules.

var con = require('./lib/console');
var tun = require('./lib/tunnel');
var srv = require('./lib/server');

// Set up variables pointing to our config directory, certificate files and
// subdirectories for tunnels and packages.

var configDir = path.join(process.env.HOME, '.mole');
var configFile = path.join(configDir, 'mole.ini');
var certFile = path.join(configDir, 'mole.crt');
var keyFile = path.join(configDir, 'mole.key');
var tunnelDefDir = path.join(configDir, 'tunnels');
var pkgDir = path.join(configDir, 'pkg');

// Create the tunnel and package directories. Any needed components leading up
// to these directories will be created as well as needed. No harm if they
// already exist.

mkdirp.sync(tunnelDefDir);
mkdirp.sync(pkgDir);

// Mark the entire config directory as private since we'll be storing keys and
// passwords in plaintext in there.

fs.chmodSync(configDir, 448 /* 0700 octal */);

// Load the config file. If it doesn't exist, set defaults and write a new
// config file.

var config = new inireader.IniReader();
try {
    config.load(configFile);
} catch (err) {
    con.info('No config, using defaults.');
    config.param('server.port', 9443);
    config.write();
}

// Set up the help text that will be appended after the commands and options
// summary when the user makes an error or runs mole without parameters.

var helptext = [
      'Version:',
      '  mole v' + pkg.version + '\t(https://github.com/calmh/mole)',
      '  node ' + process.version,
      '',
      'Examples:',
      '',
      'Register with server "mole.example.com" and a token:',
      '  mole register mole.example.com 80721953-b4f2-450e-aaf4-a1c0c7599ec2'.bold,
      '',
      'List available tunnels:',
      '  mole list'.bold,
      '',
      'Dig a tunnel to "operator3":',
      '  mole dig operator3'.bold,
      '',
      'Fetch new and updated tunnel specifications from the server:',
      '  mole pull'.bold
].join('\n');

// Set the name of our 'script'.

parser.script('mole');

// `dig <destination> [host]`

parser.command('dig')
.help('Dig a tunnel to the destination')
.option('tunnel', { position: 1, help: 'Tunnel name or file name', required: true })
.option('host', { position: 2, help: 'Host name within tunnel definition' })
.callback(dig);

// `list`

parser.command('list')
.help('List available tunnel definitions')
.callback(list);

// `pull`

parser.command('pull')
.help('Get tunnel definitions from the server')
.callback(pull);

// `push <file>`

parser.command('push')
.help('Send a tunnel definition to the server')
.option('file', { position: 1, help: 'File name', required: true })
.callback(push);

// `register <host> <token>`

parser.command('register')
.help('Register with a mole server')
.option('server', { position: 1, help: 'Server name', required: true })
.option('token', { position: 2, help: 'One time registration token', required: true })
.option('port', { abbr: 'p', metafile: 'PORT', help: 'Server port (default 9443)', default: 9443 })
.callback(register);

// `gettoken`

parser.command('gettoken')
.help('Generate a new registration token')
.callback(token);

// `export <tunnel> <file>`

parser.command('export')
.option('tunnel', { position: 1, help: 'Tunnel name', required: true })
.option('file', { position: 2, help: 'File name to write tunnel definition to', required: true })
.help('Export tunnel definition to a file')
.callback(exportf);

// `newuser [-a] <name>`

parser.command('newuser')
.option('name', { position: 1, help: 'User name', required: true })
.option('admin', { flag: true, abbr: 'a', help: 'Create an admin user' })
.help('Create a new user (requires admin privileges)')
.callback(newUser);

// `deluser <name>`

parser.command('deluser')
.option('name', { position: 1, help: 'User name', required: true })
.help('Delete a user (requires admin privileges)')
.callback(delUser);

// `install <pkg>`

parser.command('install')
.option('pkg', { position: 1, help: 'Package name', required: true })
.help('Install an optional package, fetched from the server')
.callback(install);

// `-d` always turns on debug.

parser.option('debug', { abbr: 'd', flag: true, help: 'Display debug output' });

// `-h` shows help. This is actually implemented totally by `nomnom`, but we
// need to define the option so it shows up in the usage information.

parser.option('help', { abbr: 'h', flag: true, help: 'Display command help' });

// If no command was given, we print the usage information and exit. We also
// tack on the `.help` call to set the global help text here.  I'd rather do
// that in a separate call on `parser` instead of chaining after `.nocommand`
// etc, but that doesn't actually work for reasons I don't understand.

parser.nocommand()
.callback(function () {
    console.log(parser.getUsage());
    process.exit(0);
})
.help(isatty ? helptext : helptext.stripColors);

// Parse command line arguments. This will call the defined callbacks for matching commands.

parser.parse();

// Try to read our certificates and pass them to the `server` instance. Fail
// silently if there are no certificates, since we might simply not be
// registered yet.

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
    }
}

// Initialize stuff given the options from `nomnom`. This must be called early from every command callback.

function init(opts) {
    if (opts.debug) {
        con.enableDebug();
    }

    // The server code will check the certificate fingerprint if we have one
    // stored from before.  If not, we'll store the fingerprint on `register`,
    // thus effectively locking the client to the server it registered with and
    // preventing some tampering scenarios.
 
    srv.init({ host: config.param('server.host'), port:
             config.param('server.port') || 9443, fingerprint:
             config.param('server.fingerprint') });

    readCert();
}

function register(opts) {
    init(opts);

    con.debug('Requesting registration from server ' + opts.server + ':' + opts.port);

    // Set the server and port we received from parameters in the config file,
    // and tell the server code to use them.  We don't save the config just
    // yet, though.

    config.param('server.host', opts.server);
    config.param('server.port', opts.port);
    srv.init({ host: opts.server, port: opts.port });

    // Try to register with the server. If it fails, a fatal error will be
    // printed by the server code and the callback will never be called.  If it
    // succeeds, we'll get our certificates and the server fingerprint in the
    // callback.

    srv.register(opts.token, function (result) {
        con.debug('Received certificate and key from server');

        // Save the certificates and fingerprint for later.

        fs.writeFileSync(certFile, result.cert);
        fs.writeFileSync(keyFile, result.key);
        config.param('server.fingerprint', result.fingerprint);

        // Save the config file since we've verified the server and port and
        // got the fingerprint.

        config.write();
        con.ok('Registered');

        // Read our newly minted certificates and do a `pull` to get tunnel
        // definitions.

        readCert();
        pull(opts);
    });
}

function token(opts) {
    init(opts);

    con.debug('Requesting new token from server');
    con.info('A token can be used only once');
    con.info('Only the most recently generated token is valid');

    // Request a token from the server. On success, the callback will be called
    // with the token and we simply print it out.

    srv.token(function (result) {
        con.ok(result.token);
    });
}

function list(opts) {
    init(opts);

    // Get a sorted list of all files in the tunnel directory.

    con.debug('listing files in ' + tunnelDefDir);
    var files = fs.readdirSync(tunnelDefDir);
    files.sort();
    con.debug('Got ' + files.length + ' files');

    // Build a table with information about the tunnel definitions. Basically,
    // load each of them, create a row with information and push that row to
    // the table.

    var rows = [];
    files.forEach(function (file) {
        var tname = tun.name(file);
        try {
            var r = tun.load(path.join(tunnelDefDir, file));

            var opts = '';
            if (r.vpnc) {
                opts += ' (vpn)'.magenta;
            }

            var descr = r.description;
            // FIXME: For lots of hosts, this isn't all that useful since it'll be truncated by the table formatter.
            var hosts = _.keys(r.hosts).sort().join(', ');
            var mtime = r.stat.mtime;
            var mdate = iso8601.fromDate(mtime).slice(0, 10)
            rows.push([ tname.blue.bold , descr + opts, mdate, hosts ]);
        } catch (err) {
            // If we couldn't load/parse the file for some reason, simply mark it as corrupt.
            rows.push([ tname.red.bold, '--Corrupt--', '--Corrupt--', '--Corrupt--' ]);
        }
    });

    // Format the table using the specified headers and the rows from above.

    table([ 'TUNNEL', 'DESCRIPTION', 'MODIFIED', 'HOSTS' ], rows, { underlineHeaders: true });
}

function pull(opts) {
    init(opts);

    // Get the list of tunnel definitions from the server. The list includes
    // (name, mtime) for each tunnel. We'll use the `mtime` to figure out if we
    // need to download the definition or not.

    con.debug('Requesting tunnel list from server');
    srv.list(function (result) {
        con.debug('Got ' + result.length + ' entries');
        con.debug(util.inspect(result));

        // We use this to keep track of the number of outstanding requests and
        // to print a message when every request has finished.

        var inProgress = 0;
        function done() {
            inProgress -= 1;
            if (inProgress === 0) {
                con.ok(result.length + ' tunnel definitions in sync');
                inProgress = -1;
            }
        }

        result.forEach(function (res) {
            inProgress += 1;

            // Figure out the local filename that corresponds to this tunnel, if we have it.

            var local = path.join(tunnelDefDir, res.name);

            // We need to fetch the file only if we either don't already have
            // it, or if the mtime as sent by the server differs from what we
            // have locally.

            var fetch = false;
            if (!existsSync(local)) {
                fetch = true;
            } else {
                var s = fs.statSync(local);
                if (s.mtime.getTime() !== parseInt(res.mtime, 10)) {
                    fetch = true;
                }
            }

            // If we need to fetch a tunnel definition, send a server request to do so.

            if (fetch) {
                srv.fetch(res.name, function (result) {

                    // When the request completes, we save the file and set the
                    // mtime to match that sent by the server. The server sends
                    // it in milliseconds since that's what Javascript
                    // timestamps usually are, but utimesSync expects seconds.

                    fs.writeFileSync(local, result);

                    var mtime = Math.floor(res.mtime / 1000);
                    fs.utimesSync(local, mtime, mtime);

                    con.ok('Pulled ' + tun.name(res.name));

                    // Mark this request as completed, print out status if it was the last one.

                    done();
                });
            } else {

                // Mark this request as completed. We don't do it immediately
                // since that would result in the `inProgress` counter flapping
                // between 1 and 0 when we didn't need to fetch anything.
                // Instead, queue the call for the next tick, when `inProgress`
                // has been incremented all the way.

                process.nextTick(done);
            }
        });
    });

    // The user presumably want's to be up to date, so we fetch the latest
    // version number for mole from the npm repository and print a 'time to
    // upgrade'-message if there's a mismatch.
   
    version.fetch('mole', function (err, ver) {
        if (!err && ver) {
            if (ver !== pkg.version) {
                con.info('You are using mole v' + pkg.version + '; the latest version is v' + ver);
                con.info('Use "sudo npm -g update mole" to upgrade');
            } else {
                con.ok('You are using the latest version of mole');
            }
        }
    });
}

function push(opts) {
    init(opts);

    // We load the tunnel, which will cause some validation of it to happen. We
    // don't want to push files that are completely broken.

    con.debug('Testing ' + opts.file);
    try {
        tun.load(opts.file);
        con.debug('It passed validation');
    } catch (err) {
        con.fatal(err);
    }

    // We read the file to a buffer to send to the server. There should be no
    // errors here since the tunne load and check above succeeded.

    con.debug('Reading ' + opts.file);
    var data = fs.readFileSync(opts.file, 'utf-8');

    // Send the data to the server. We'll only get the callback if the upload
    // succeeds.

    srv.send(path.basename(opts.file), data, function (result) {
        con.ok('Sent ' + data.length + ' bytes');
    });
}

function newUser(opts) {
    init(opts);

    // Create a new user on the server. If the call succeeds, we'll get the
    // callback with the one-time token for the new user.

    con.debug('Requesting user ' + opts.name);
    srv.newUser(opts.name, function (result) {
        con.ok(result.token);
    });
}

function delUser(opts) {
    init(opts);

    // Delete a user from the server. As always, we only get the callback if
    // everything went well.

    con.debug('Deleting user ' + opts.name);
    srv.delUser(opts.name, function (result) {
        con.ok('deleted');
    });
}

function dig(opts) {
    init(opts);

    // Before we do any digging, we make sure that `expect` is available.

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

// Here's the real meat of `mole`, the tunnel digging part.

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

    // If a specific host was requested, we modify the configuration to use
    // that host as `main`.

    if (host) {
        con.debug('Using specified main host ' + host);
        config.main = host;
    }

    // Create and save the ssh config. We add some defaults to the top to avoid
    // complaints about missing or mismatching host keys.

    con.debug('Creating ssh configuration');
    var defaults = [
        'Host *',
        '  UserKnownHostsFile /dev/null',
        '  StrictHostKeyChecking no'
    ].join('\n') + '\n';
    var conf = defaults + sshConfig(config);
    fs.writeFileSync(config.sshConfig, conf);
    con.debug(config.sshConfig);

    // If the tunnel definition specifies a VPN connection, we need to get that
    // up and running before we call expect.

    if (config.vpnc) {

        // First we make sure that `vpnc` is actually installed, or exit with a
        // helpful suggestion if it's not.

        vpnc.available(function (err, result) {
            if (err) {
                con.error(err);
                con.error('vpnc unavailable; try "mole install vpnc"');
                con.fatal('Not continuing without vpnc');
            } else {
                con.debug('Using ' +  result.version);

                // If vpnc is already running, it's almost certainly going to
                // fail to bring up one more VPN connection. So if that's the
                // case, exit with an error.

                pidof('vpnc', function (err, pid) {
                    if (err) {
                        con.error(err);
                        con.fatal('could not check if vpnc was running');
                    } else if (pid) {
                        con.fatal('vpnc already running, will not start another instance');
                    }

                    // Try to connect the VPN. If it fails, exit with an error,
                    // otherwise proceed to start expect and to the real SSH
                    // tunnelling.

                    con.info('Connecting VPN; you might be asked for your local (sudo) password now');
                    vpnc.connect(config.vpnc, config.vpnRoutes, function (err, code) {
                        if (err) {
                            con.fatal(err);
                        } else if (code !== 0) {
                            con.fatal('vpnc returned an error - investigate and act on it, nothing more I can do :(');
                        }
                        con.info('VPN connected. Should the login sequence fail, you can disconnect the VPN');
                        con.info('manually by running "sudo ' + result.vpncDisconnect + '"');
                        launchExpect(config, debug);
                    });
                });
            }
        });
    } else {

        // We don't need a VPN connection, so go directly to the expect step.

        launchExpect(config, debug);
    }
}

function launchExpect(config, debug) {
    var setupLocalIPs = require('./lib/setup-local-ips');
    var expectConfig = require('./lib/expect-config');

    // Create and save the expect script that's going to drive the session.

    con.debug('Creating expect script');
    try {
        var expect = expectConfig(config, debug);
        var expectFile = temp.path({suffix: '.expect'});
        fs.writeFileSync(expectFile, expect);
        con.debug(expectFile);
    } catch (err) {
        con.fatal(err);
    }

    // Set up local IP:s needed for forwarding. If it fails, we remove the
    // forwardings from the configuration and warn the user, but proceed
    // anyway.

    con.debug('Setting up local IP:s for forwarding');
    setupLocalIPs(config, function (c) {
        if (!c) {
            con.warning('Failed to set up IP:s for forwarding. Continuing without forwarding.');
            delete config.forwards;
        }

        // Start the expect script and wait for it to exit. We use the
        // deprecated `customFds` option to connect the script to the tty so
        // the user can interact with it. When we migrate to Node 0.8, there's
        // a supported `stdio: inherit` option we can use instead.

        con.info('Hang on, digging the tunnel');
        spawn('expect', [ expectFile ], { customFds: [ 0, 1, 2 ] })
        .on('exit', function (code) {

            // The script has exited, so we try to clean up after us.

            fs.unlinkSync(expectFile);
            fs.unlinkSync(config.sshConfig);
            // FIXME: Unlink ssh keys

            // If a VPN was connected, now is the time to disconnect it. We
            // don't treat errors here as fatal since there could be more
            // cleanup to do later on and we're anyway exiting soon.

            if (config.vpnc) {
                con.info('Disconnecting VPN; you might be asked for your local (sudo) password now');
                vpnc.disconnect(function (err, status) {
                    if (err) {
                        con.error(err);
                        con.ok('VPN disconnection failed');
                    } else {
                        con.ok('VPN disconnected');
                    }
                });
            }

            // Print final status message. If things seems to have failed,
            // suggest turning on debugging or talking to the author of the
            // tunnel definition.

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

    // Load and verify the tunnel.

    try {
        con.debug('Loading tunnel');
        config = tun.load(opts.tunnel, tunnelDefDir);
    } catch (err) {
        con.fatal(err);
    }

    // Save it out to the specified file.

    con.debug('Saving to INI format');
    tun.save(config, opts.file);

    con.ok(opts.file);
}

function install(opts) {
    init(opts);

    // We build the expected package name based on the name specified by the
    // user, plus the platform, architecture and OS version.  FIXME: The OS
    // version is way too specific.

    var file = [ opts.pkg, os.platform(), os.arch(), os.release() ].join('-') + '.tar.gz';
    var local = path.join(pkgDir, file);

    // Get the package from the server and save it in our package directory.
    // The callback will be called only if the fetch and save is successfull.

    con.info('Fetching ' + file);
    srv.saveBin('/extra/' + file, local, function () {

        // Create a temporary path where we can extract the package.

        var tmp = temp.path();
        mkdirp(tmp);

        // Change working directory to the temporary one we created and try to
        // extract the downloaded package file.

        con.info('Unpacking ' + file);
        exec('cd ' + tmp + ' && tar zxf ' + local, function (err, stdout, stderr) {
            con.debug('Extracted in ' + tmp);

            // The package should include a script `install.sh` that will do
            // whatever's necessary to install the package. We run that with
            // sudo.

            con.info('Running installation, you might now be asked for your local (sudo) password.');
            var inst = spawn('sudo', [ path.join(tmp, 'install.sh'), tmp ], { customFds: [ 0, 1, 2 ] });
            inst.on('exit', function (code) {

                // We're done, one way or the other.

                if (code === 0) {
                    con.ok('Installation complete');
                } else {
                    con.info('Installation failed. Sorry.');
                }
            });
        });
    });
}

