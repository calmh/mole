"use strict";

var _ = require('underscore');
var exec = require('child_process').exec;
var fs = require('fs');
var pidof = require('pidof');
var readline = require('readline');
var temp = require('temp');
var vpnc = require('vpnc');

var con = require('../lib/console');
var expectConfig = require('../lib/expect-config');
var Proxy = require('../lib/trivial-proxy');
var pspawn = require('../lib/pspawn');
var setupLocalIPs = require('../lib/setup-local-ips');
var sshConfig = require('../lib/ssh-config');
var tun = require('../lib/tunnel');

module.exports = dig;
dig.help = 'Dig a tunnel to the destination';
dig.options = {
    local: { abbr: 'l', help: 'Tunnel name is a local file', flag: true },
    tunnel: { position: 1, help: 'Tunnel name or file name', required: true },
    host: { position: 2, help: 'Host name within tunnel definition' },
};
dig.prio = 1;

function dig(opts, state) {
    // Before we do any digging, we make sure that `expect` is available.

    exec('expect -v', function (error, stdout, stderr) {
        if (error) {
            con.error(error.toString().trim());
            con.error('Verify that "expect" is installed and available in the path.');
        } else {
            con.debug(stdout.trim());
            digReal(opts, state);
        }
    });
}

// Here's the real meat of `mole`, the tunnel digging part.

function digReal(opts, state) {
    var config;

    // Load a configuration, generate a temporary filename for ssh config.

    con.debug('Loading tunnel');
    try {
        if (opts.local) {
            config = tun.loadFile(opts.tunnel);
        } else {
            config = tun.loadByName(opts.tunnel, state.path.tunnels);
        }
    } catch (err) {
        con.fatal(err);
    }
    config.sshConfig = temp.path({suffix: '.ssh.conf'});

    // If a specific host was requested, we modify the configuration to use
    // that host as `main`.

    if (opts.host) {
        con.debug('Using specified main host ' + opts.host);
        config.main = opts.host;
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
                        con.warning('vpnc already running; consider disconnecting the VPN manually by running:');
                        con.warning('sudo ' + result.vpncDisconnect);
                        con.fatal('Not continuing');
                    }

                    // Try to connect the VPN. If it fails, exit with an error,
                    // otherwise proceed to start expect and to the real tunnelling.

                    con.info('Connecting VPN; you might be asked for your local (sudo) password now');
                    vpnc.connect(config.vpnc, config.vpnRoutes, function (err, code) {
                        if (err) {
                            con.fatal(err);
                        } else if (code !== 0) {
                            con.fatal("vpnc returned an error - investigate and act on it, nothing more I can do :(");
                        }
                        con.info('VPN connected. Should the login sequence fail, you can disconnect the VPN');
                        con.info('manually by running:');
                        con.info('sudo ' + result.vpncDisconnect);

                        setupIPs(config, opts.debug);
                    });
                });
            }
        });
    } else {

        // We don't need a VPN connection, so go directly to the next step.

        setupIPs(config, opts.debug);
    }
}

function setupIPs(config, debug) {
    // Set up local IP:s needed for forwarding. If it fails, we remove the
    // forwardings from the configuration and warn the user, but proceed
    // anyway.

    con.debug('Setting up local IP:s for forwarding');
    setupLocalIPs(config, function (c) {
        if (!c) {
            con.warning('Failed to set up IP:s for forwarding. Continuing without forwarding.');
            delete config.forwards;
        }

        if (config.main) {
            con.debug('There is a main host, going to expect');
            launchExpect(config, debug);
        } else if (_.size(config.localForwards) > 0) {
            setupLocalForwards(config);
        } else {
            con.fatal('Got this far, but now what?');
        }
    });
}

function setupLocalForwards(config) {
    var forwards = [];
    var rl = readline.createInterface(process.stdin, process.stdout);

    console.log('\nThe following forwards are available to you:\n');
    _.each(config.localForwards, function (fs, descr) {
        console.log(descr);
        fs.forEach(function (f) {
            var from = f.from.split(':');
            var to = f.to.split(':');
            forwards.push(new Proxy(from[0], parseInt(from[1], 10), to[0], parseInt(to[1], 10)));
            console.log('   ' + f.from + ' -> ' + f.to);
        });
        console.log();
    });

    // Start a very simple command loop to wait for when the user is done.

    con.ok('Go forth and connect. Type "quit" when you\'re done.');
    rl.setPrompt('mole> ');
    rl.prompt();
    rl.on('line', function (cmd) {
        if (cmd === 'quit') {
            con.ok('All done, shutting down.');

            // Shut down each port forward.

            forwards.forEach(function (f) {
                f.end();
            });

            // When the CLI has shut down, stop the VPN.

            rl.on('close', function () {
                // FIXME: Why is process.exit still necessary after rl.close?

                stopVPN(config, process.exit);
            });

            // Shut down the CLI.

            rl.close();
        } else {
            if (cmd !== '') {
                con.error('Invalid command "' + cmd + '"');
            }
            rl.prompt();
        }
    });
}

function launchExpect(config, debug) {
    var expectFile = temp.path({suffix: '.expect'});

    // Create and save the expect script that's going to drive the session.

    con.debug('Creating expect script');
    try {
        var expect = expectConfig(config, debug);
        fs.writeFileSync(expectFile, expect);
        con.debug(expectFile);
    } catch (err) {
        con.fatal(err);
    }

    // Start the expect script and wait for it to exit. We use the
    // deprecated `customFds` option to connect the script to the tty so
    // the user can interact with it. When we migrate to Node 0.8, there's
    // a supported `stdio: inherit` option we can use instead.

    con.info('Hang on, digging the tunnel');
    pspawn('expect', [ expectFile ]).on('exit', function (code) {

        // The script has exited, so we try to clean up after us.

        fs.unlinkSync(expectFile);
        fs.unlinkSync(config.sshConfig);
        // FIXME: Unlink ssh keys

        stopVPN(config, function (code) {

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

function stopVPN(config, callback) {
    // If a VPN was connected, now is the time to disconnect it. We
    // don't treat errors here as fatal since there could be more
    // cleanup to do later on and we're anyway exiting soon.

    if (config.vpnc) {
        con.info('Disconnecting VPN; you might be asked for your local (sudo) password now');
        vpnc.disconnect(function (err, status) {
            if (err) {
                con.error(err);
                con.warning('VPN disconnection failed');
            } else if (status !== 0) {
                con.warning('VPN disconnection failed (vpnc/sudo exit code ' + status + ')');
            } else {
                con.ok('VPN disconnected');
            }

            if (callback && typeof callback === 'function') {
                callback(status === 0);
            }
        });
    }
}
