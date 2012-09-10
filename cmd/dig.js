"use strict";

var _ = require('underscore');
var debuggable = require('debuggable');
var exec = require('child_process').exec;
var fs = require('fs');
var pidof = require('pidof');
var readline = require('readline');
var spawn = require('child_process').spawn;
var temp = require('temp');

var con = require('../lib/console');
var expectConfig = require('../lib/expect-config');
var Proxy = require('../lib/trivial-proxy');
var localIPs = require('../lib/setup-local-ips');
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
dig.aliases = [ 'conn', 'connect' ];
debuggable(dig);

var vpnProviders = {};
vpnProviders.vpnc = require('vpnc');
dig.dforward(vpnProviders.vpnc);
vpnProviders.openconnect = require('openconnect');
dig.dforward(vpnProviders.openconnect);

function dig(opts, state) {
    // Before we do any digging, we make sure that `expect` is available.

    exec('expect -v', function checkExpect(error, stdout, stderr) {
        if (error) {
            con.error(error.toString().trim());
            con.error('Verify that "expect" is installed and available in the path.');
        } else {
            dig.dlog('found expect', {output: stdout.trim()});
            digReal(opts, state);
        }
    });
}

// Here's the real meat of `mole`, the tunnel digging part.

var connectedVpn;
function digReal(opts, state) {
    var config;

    // Load a configuration, generate a temporary filename for ssh config.

    dig.dlog('Loading tunnel');
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
        dig.dlog('Using specified main host ' + opts.host);
        config.main = opts.host;
    }

    // Create and save the ssh config. We add some defaults to the top to avoid
    // complaints about missing or mismatching host keys.

    dig.dlog('Creating ssh configuration');
    var defaults = [
        'Host *',
        '  UserKnownHostsFile /dev/null',
        '  StrictHostKeyChecking no'
    ].join('\n') + '\n';
    var conf = defaults + sshConfig(config, opts.debug);
    fs.writeFileSync(config.sshConfig, conf);
    dig.dlog(config.sshConfig);

    // If the tunnel definition specifies a VPN connection, we need to get that
    // up and running before we call expect.

    var handled = false;
    _.each(vpnProviders, function (provider, name) {
        if (!config[name]) {
            return;
        }

        handled = true;

        if (opts.debug && provider.setDebug) {
            dig.dlog('Enabling debug for ' + name);
            provider.setDebug();
        }

        // First we make sure that the required VPN provider is actually
        // installed, or exit with a helpful suggestion if it's not.

        provider.available(function (err, result) {
            if (err) {
                con.error(err);
                con.error(name + ' unavailable; try "mole install ' + name + '"');
                con.fatal('Not continuing without ' + name);
            } else {
                dig.dlog('Using ' +  result.version);

                // If the VPN provider is already running, it's almost
                // certainly going to fail to bring up one more VPN connection.
                // So if that's the case, exit with an error.

                pidof(name, function (err, pid) {
                    if (err) {
                        con.error(err);
                        con.fatal('Could not check if ' + name + ' was running');
                    } else if (pid) {
                        con.fatal(name + ' already running; disconnect the VPN manually');
                    }

                    // Try to connect the VPN. If it fails, exit with an error,
                    // otherwise proceed to start expect and to the real tunnelling.

                    con.info('Connecting VPN; you might be asked for your local (sudo) password now');
                    provider.connect(config[name], config.vpnRoutes, function (err, code) {
                        if (err) {
                            con.fatal(err);
                        } else if (code !== 0) {
                            con.error(name + ' returned an error (code ' + code + ')');
                            if (opts.debug) {
                                con.fatal('Inspect any output above for clues...');
                            } else {
                                con.fatal('Retry with `mole dig -d <destination>` for more information.');
                            }
                        }
                        con.info('VPN connected.');

                        connectedVpn = provider;
                        setupIPs(config, opts.debug);
                    });
                });
            }
        });
    });

    if (!handled) {
        // We don't need a VPN connection, so go directly to the next step.
        setupIPs(config, opts.debug);
    }
}

function setupIPs(config, debug) {
    // Set up local IP:s needed for forwarding. If it fails, we remove the
    // forwardings from the configuration and warn the user, but proceed
    // anyway.

    dig.dlog('Setting up local IP:s for forwarding');
    localIPs.add(config, function (err) {
        if (err) {
            con.warning('Failed to set up IP:s for forwarding. Continuing without forwarding.');
            delete config.forwards;
        }

        if (config.main) {
            dig.dlog('There is a main host, going to expect');
            launchExpect(config, debug);
        } else if (_.size(config.forwards) > 0) {
            setupLocalForwards(config);
        } else {
            con.fatal('Got this far, but now what?');
        }
    });
}

function removeIPs(config, debug) {
    dig.dlog('Removing extra local IP:s');
    localIPs.remove(config, function (err) {
        if (err) {
            con.warning('Failed to remove IP:s.');
        }
    });
}

function setupLocalForwards(config) {
    var forwards = [];
    var rl = readline.createInterface(process.stdin, process.stdout);

    console.log('\nThe following forwards are available to you:\n');
    _.each(config.forwards, function (fs, descr) {
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

    con.ok('Go forth and connect. Type "quit" or ^D when you\'re done.');
    rl.setPrompt('mole> ');
    rl.prompt();

    // Read commands.
    rl.on('line', function (cmd) {
        if (cmd === 'quit') {
            // Shut down the CLI.
            rl.close();
        } else {
            if (cmd !== '') {
                con.error('Invalid command "' + cmd + '"');
            }
            rl.prompt();
        }
    });

    // When the CLI has shut down, stop the VPN.
    rl.on('close', function () {
        con.ok('All done, shutting down.');

        // Shut down each port forward.

        forwards.forEach(function (f) {
            f.end();
        });

        // 'close' is emitted before the close is completed.  Schedule
        // the stop on the next tick, to allow readline to clean up the
        // console etc.

        process.nextTick(function () {
            removeIPs(config);
            stopVPN(config, function (code) {
                finalExit(code, config);
            });
        });
    });

}

function launchExpect(config, debug) {
    var expectFile = temp.path({suffix: '.expect'});

    // Create and save the expect script that's going to drive the session.

    dig.dlog('Creating expect script');
    try {
        var expect = expectConfig(config, debug);
        fs.writeFileSync(expectFile, expect);
        dig.dlog(expectFile);
    } catch (err) {
        con.fatal(err);
    }

    // Start the expect script and wait for it to exit. We use the
    // deprecated `customFds` option to connect the script to the tty so
    // the user can interact with it. When we migrate to Node 0.8, there's
    // a supported `stdio: inherit` option we can use instead.

    con.info('Hang on, digging the tunnel');
    spawn('expect', [expectFile], {stdio: 'inherit'}).on('exit', function (expCode) {
        dig.dlog('expect complete', {exit: expCode});

        // The script has exited, so we try to clean up after us.

        fs.unlinkSync(expectFile);
        fs.unlinkSync(config.sshConfig);
        // FIXME: Unlink ssh keys

        removeIPs(config, debug);
        stopVPN(config, function (code) {
            finalExit(code + expCode, config);
        });
    });
}

function finalExit(code, config) {

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
}

function stopVPN(config, callback) {
    // If a VPN was connected, now is the time to disconnect it. We
    // don't treat errors here as fatal since there could be more
    // cleanup to do later on and we're anyway exiting soon.

    if (connectedVpn) {
        con.info('Disconnecting VPN; you might be asked for your local (sudo) password now');
        connectedVpn.disconnect(function (err, status) {
            if (err) {
                con.error(err);
                con.warning('VPN disconnection failed');
            } else if (status !== 0) {
                con.warning('VPN disconnection failed (exit code ' + status + ')');
            } else {
                con.ok('VPN disconnected');
            }

            if (callback && typeof callback === 'function') {
                callback(status);
            }
        });
    } else {
        callback(0);
    }
}
