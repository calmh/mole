#!/usr/bin/env node

// We now use strict for everything since Node supports it.

"use strict";

var _ = require('underscore');
var con = require('yacon');
var fs = require('fs');
var ini = require('ini');
var mkdirp = require('mkdirp');
var rimraf = require('rimraf');
var parser = require('nomnom');
var path = require('path');

// The existsSync function moved between Node 0.6 and 0.8. We monkeypatch fs if
// it's not already there.
//
// You'll be seeing a lot of `*Sync` calls in here. If that disturbs you, keep
// in mind that this is a CLI program that runs once and then exits, not some
// sort of high performance IO-bound server code, mmkay?

if (!fs.existsSync) {
    fs.existsSync = path.existsSync;
}

var state = {};

// We load our own package file to get at the version number.

state.pkg = require(path.join(__dirname, 'package.json'));

var init = require('./lib/init');
var Client = require('./lib/client');

state.client = new Client();

// All server errors are fatal.

state.client.on('error', function (err) {
    con.fatal(err);
});

// Make a best guess of the user's home directory, with fallback to
// /tmp/<username>.

var homeDir = path.resolve(process.env.HOME || process.env.USERPROFILE ||
                           path.join('/tmp', process.env.USER || process.env.LOGNAME));

// Set up variables pointing to our config directory, certificate files and
// subdirectories for tunnels and packages.

state.path = { configDir: path.join(homeDir, '.mole') };
state.path.certFile = path.join(state.path.configDir, 'mole.crt');
state.path.configFile = path.join(state.path.configDir, 'mole.ini');
state.path.keyFile = path.join(state.path.configDir, 'mole.key');
state.path.pkgDir = path.join(state.path.configDir, 'pkg');
state.path.tunnels = path.join(state.path.configDir, 'tunnels');

// Create the package directory. Any needed components leading up
// to these directories will be created as well as needed. No harm if they
// already exist.

mkdirp.sync(state.path.pkgDir);

// The tunnel directory is deprected in mole v3. Remove it and any files inside it.

rimraf.sync(state.path.tunnels);


// Mark the entire config directory as private since we'll be storing keys and
// passwords in plaintext in there.

fs.chmodSync(state.path.configDir, 448 /* 0700 octal */);

// Load the config file. If it doesn't exist, set defaults and write a new
// config file.

try {
    state.config = ini.parse(fs.readFileSync(state.path.configFile, 'utf-8'));
} catch (err) {
    con.info('No config, using defaults.');
    state.config = {
        server: {
            port: 9443
        }
    };
    fs.writeFileSync(state.path.configFile, ini.stringify(state.config));
}

// Set the name of our 'script'.

parser.script('mole');

// Load all command modules. Add them into an array with prio and name as the
// first elements to allow easy sorting. Sort by prio, name.

var cmds = fs.readdirSync(path.join(__dirname, 'cmd'))
.map(function (module) {
    if (!module.match(/^[a-z0-9]+\.js$/)) { return null; }

    var cmd = require('./cmd/' + module);
    var name = path.basename(module, '.js');
    return [ cmd.prio || 5, name, cmd ];
})
.sort();

// Add them to the command line parser.

cmds.forEach(function (arr) {
    if (!arr) { return; }

    var name = arr[1];
    var module = arr[2];
    var names = [ name ];
    if (module.aliases) {
        names = names.concat(module.aliases);
    }

    names.forEach(function (name) {
        var cmdp = parser.command(name);

        cmdp.help(module.help);

        _.each(module.options, function (v, k) {
            cmdp.option(k, v);
        });

        cmdp.callback(function (opts) {
            init(opts, state);
            module(opts, state);
        });
    });
});

// `-d` always turns on debug.
parser.option('debug', { abbr: 'd', flag: true, help: 'Display debug output' });

// `-h` shows help. This is actually implemented totally by `nomnom`, but we
// need to define the option so it shows up in the usage information.
parser.option('help', { abbr: 'h', flag: true, help: 'Display command help' });

// `-r` permit uid 0 execution. Allows the user to execute mole as root.
parser.option('permit-root', { abbr: 'r', flag: true, help: 'Permit root execution' });

// Generate the help text.
function usage(cmds) {
    var sprint = require('sprint');

    var str = '';
    str += '\nUsage: mole <command> [options]\n';

    str += '\nCommands:\n';
    cmds.forEach(function (c) {
        str += sprint('   %-12s', c[1]).bold + ' ' + c[2].help;
        if (c[2].aliases) {
            str += ' (alias: ' + c[2].aliases.join(', ').bold + ')';
        }
        str += '\n';
    });

    str += '\nOptions:\n';
    [
        [ '-d, --debug', 'Display debug output' ],
        [ '-h, --help', 'Display command help' ],
        [ '-r, --permit-root', 'Permit root execution' ]
    ].forEach(function (o) {
        str += sprint('   %-12s', o[0]).bold + ' ' + o[1] + '\n';
    });
    str += '\n';
    str += '   Use ' + 'mole <command> -h'.bold + ' to see command specific options.\n';
    str += '\n';

    str += [
        'Version:',
        '   mole v' + state.pkg.version + '\t(https://github.com/calmh/mole)',
        '   node ' + process.version,
        '',
        'Examples:',
        '',
        '   Register with server "mole.example.com" and a token:',
        '   mole register mole.example.com 80721953-b4f2-450e-aaf4-a1c0c7599ec2'.bold,
        '',
        '   List available tunnels:',
        '   mole list'.bold,
        '',
        '   Dig a tunnel to "operator3":',
        '   mole dig operator3'.bold
    ].join('\n');

    str += '\n';
    return str;
}

var args = process.argv.slice(2);
if (args.length === 0 || args[0] === '-h' || args[0] === '--help') {
    // Print usage
    console.log(usage(cmds));
} else {
  // Parse command line arguments. This will call the defined callbacks for matching commands.
  var opts = parser.parse();

  // Prevent running as root.
  if (process.getuid && process.getuid() === 0 && !opts['permit-root']) {
    con.fatal('Running mole as root is generally unnecessary, please provide the -r flag to force execution.');
  }
}

