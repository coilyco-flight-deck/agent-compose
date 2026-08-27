"""Checks that exist only while agent-compose still holds the Go engine.

The differential tests here are agent-compose#333's oracle: housecast composes
and this repository proves the Go engine agrees, byte for byte. The palette probe
reads the band constants out of internal/color/color.go for the same reason.

All of it is deleted with the Go semantic layer under agent-compose#339.
"""
