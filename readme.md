===============================================================================
POWER FAILURE RECOVERY DAEMON
===============================================================================

WHAT IS THIS?

This is a program that runs in the background on your Linux laptop and 
automatically saves what you're doing. If your laptop suddenly dies (power 
cut, battery fails, you unplug the charger), you can restore your work 
after rebooting instead of losing everything.

THE PROBLEM IT SOLVES

Your laptop battery is damaged. When the charger disconnects, the laptop 
dies instantly. You lose:
- Unsaved files you were typing
- Browser tabs you had open
- Terminal commands you were running
- Editor sessions with unsaved changes
- Running processes

This program saves your state every 60 seconds so you never lose more than
one minute of work.

===============================================================================
HOW IT WORKS (SIMPLE EXPLANATION)
===============================================================================

The program runs quietly in the background. Every 60 seconds it:
1. Checks your battery percentage
2. Sees if the charger is plugged in
3. Looks at what programs are running
4. Saves this information to your hard drive

When you reboot after a crash, you run the program with --restore and it
tells you what you had open before the crash.

WHAT HAPPENS WHEN CHARGER DISCONNECTS

If you unplug the charger while the laptop is running, the program:
1. Detects this immediately (within 1 second)
2. Saves an emergency checkpoint
3. Continues monitoring battery drain

If battery drains faster than normal, it saves more often:
- Normal drain (slow): Save every 60 seconds
- Fast drain (15% per minute): Save every 10 seconds
- Critical battery (below 5%): Save every 1 second

===============================================================================
WHAT IT ACTUALLY SAVES
===============================================================================

The program creates a checkpoint file (JSON format) containing:

-------------------------------------------------------------------------------
1. BASIC INFORMATION
-------------------------------------------------------------------------------

- Timestamp: When the checkpoint was saved
- Save reason: Why it saved (periodic, AC loss, low battery, shutdown)
- Version: Format version number

Example:
  "timestamp": "2026-05-17T15:30:00Z"
  "save_reason": "periodic"
  "version": 1

-------------------------------------------------------------------------------
2. BATTERY STATUS
-------------------------------------------------------------------------------

- Battery percentage: How much charge remains
- Battery voltage: Current voltage in millivolts
- Discharge rate: How fast battery is draining (% per minute)
- AC online: Whether charger is plugged in

Example:
  "battery_percent": 85.2
  "battery_voltage_mv": 11800
  "discharge_rate_per_min": 2.5
  "ac_online": true

Why this matters: Tells you if the crash was caused by low battery or sudden
power loss.

-------------------------------------------------------------------------------
3. RUNNING PROCESSES
-------------------------------------------------------------------------------

For EACH running process, the program saves:

- PID: Process ID number
- Name: Process name (like "code", "firefox", "bash")
- Command: Full command line that started it
- CWD: Current working directory (which folder it was in)

Example for a single process:
  {
    "pid": 1234,
    "name": "code",
    "command": "/usr/bin/code /home/user/myproject",
    "cwd": "/home/user/myproject"
  }

This means after a crash, you know:
- VSCode was open (name: "code")
- It was editing a project in /home/user/myproject
- The full command to reopen it

-------------------------------------------------------------------------------
4. ACTIVE WINDOWS (OPTIONAL, REQUIRES xdotool)
-------------------------------------------------------------------------------

If you have the "xdotool" program installed, it also saves:

- Window title: What the window says at the top
- Window class: Type of window (like "Firefox", "Code")
- PID: Process ID that owns the window

Example:
  "active_windows": [
    {"title": "main.go - VSCode", "class": "Code", "pid": 1234},
    {"title": "GitHub - Firefox", "class": "Firefox", "pid": 5678}
  ]

===============================================================================
WHAT IT DOES NOT SAVE (YET)
===============================================================================

The following are NOT saved in the current version:

- Cursor position inside editors
- Unsaved text you were typing (only the fact that the editor was open)
- Browser tab URLs (browsers restore these themselves)
- Terminal scrollback history (only the working directory)
- Network connections
- Audio/video playback position
- Game progress

These may be added in future versions.

===============================================================================
HOW MUCH DATA DOES IT STORE?
===============================================================================

Each checkpoint is typically 10KB to 100KB depending on how many processes
are running. With the default setting of keeping 10 checkpoints, total
storage used is 1-2 MB. This is very small.

Checkpoints are stored in:
/var/lib/power-failure-recovery/

===============================================================================
CONFIGURATION OPTIONS
===============================================================================

You can change how the program behaves with these options:

--interval 60              Save every 60 seconds (default)
--interval 30              Save every 30 seconds (more frequent)
--interval 120             Save every 120 seconds (less frequent)

--aggressive-interval 10   Save every 10 seconds when battery drains fast
--emergency-interval 1     Save every 1 second when battery is critical

--min-battery 5            Emergency mode when battery below 5% (default)
--min-battery 10           Emergency mode when battery below 10% (safer)

--fast-discharge 15        Fast drain if losing 15% per minute (default)
--fast-discharge 20        Fast drain if losing 20% per minute (less sensitive)

--data-dir /custom/path    Where to save checkpoints
--max-checkpoints 10       Keep only the 10 most recent checkpoints

--dry-run                  Test without actually saving anything
--debug                    Show detailed log messages
--restore                  Restore previous session after a crash

===============================================================================
EXAMPLE USAGE
===============================================================================

Normal usage (save every 60 seconds):
  sudo ./power-recovery

Save every 30 seconds (for important work):
  sudo ./power-recovery --interval 30

Very aggressive saving (for critical work):
  sudo ./power-recovery --interval 10 --aggressive-interval 5 --min-battery 10

Test without saving:
  sudo ./power-recovery --dry-run --debug

Restore after crash:
  ./power-recovery --restore

Run as system service:
  sudo make install
  sudo systemctl start power-failure-recovery
  sudo systemctl enable power-failure-recovery

===============================================================================
WHAT YOU GET AFTER A CRASH
===============================================================================

When your laptop dies and you reboot, here's what happens:

1. Run: ./power-recovery --restore

2. The program shows:
   ========================================
   Restoring session from 2026-05-17 15:30:00
   Battery at crash: 4.8%
   Processes to restore: 47
   ========================================
     Would restore: code (VSCode)
     Would restore: firefox (Browser)
     Would restore: bash (Terminal)
   ========================================

3. You know exactly what was running before the crash

4. You can manually reopen those applications

===============================================================================
REALISTIC EXPECTATIONS
===============================================================================

WITHOUT this program (normal Linux):
- Lose everything you were doing
- Lose unsaved work
- Lose browser tabs
- Spend 30 minutes rebuilding your environment

WITH this program:
- Lose at most 60 seconds of work (or less if battery was low)
- Keep list of all running programs
- Know which directories terminals were in
- Know which projects editors had open
- Restore within 2-3 minutes

===============================================================================
LIMITATIONS (BEING HONEST)
===============================================================================

1. If your laptop dies instantly (no warning), you lose up to 60 seconds
   of work. This program cannot predict the future.

2. It cannot recover unsaved file content. It only knows that you had
   an editor open with a specific file.

3. It cannot restore browser tabs directly (but browsers have built-in
   session restore that usually works).

4. It cannot restore terminal scrollback history (only the working
   directory and shell history files).

5. It does not save cursor position inside editors.

