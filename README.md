Logitech G102 and G203 Prodigy Mouse LED control
================================================


Usage
=====

Solid color mode
----------------
`gled solid <color>`
                       
Cycle through all colors  
------------------------                     
`gled cycle <rate> <brightness>`

Single color breathing
----------------------
`gled breathe <color> <rate> <brightness>`

Enable/disable startup effect
-----------------------------
`gled intro <toggle>`


Arguments
=========

- color - RRGGBB (RGB hex value)
- rate - 100-60000 (Number of milliseconds. Default: 10000ms)
- brightness - 0-100 (Percentage. Default: 100%)
- toggle - on|off

Flags
=====

Debug level for libusb. Default: 0
----------------------------------
`gled -debug <0..3> ...   `                  