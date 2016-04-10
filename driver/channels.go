package driver

const ET_comedi = 0
const ET_simulation = 1

//in port 4
const ch_OBSTRUCTION = (0x300 + 23)
const ch_STOP = (0x300 + 22)
const ch_BUTTON_COMMAND1 = (0x300 + 21)
const ch_BUTTON_COMMAND2 = (0x300 + 20)
const ch_BUTTON_COMMAND3 = (0x300 + 19)
const ch_BUTTON_COMMAND4 = (0x300 + 18)
const ch_BUTTON_UP1 = (0x300 + 17)
const ch_BUTTON_UP2 = (0x300 + 16)

//in port 1
const ch_BUTTON_DOWN2 = (0x200 + 0)
const ch_BUTTON_UP3 = (0x200 + 1)
const ch_BUTTON_DOWN3 = (0x200 + 2)
const ch_BUTTON_DOWN4 = (0x200 + 3)
const ch_SENSOR_FLOOR1 = (0x200 + 4)
const ch_SENSOR_FLOOR2 = (0x200 + 5)
const ch_SENSOR_FLOOR3 = (0x200 + 6)
const ch_SENSOR_FLOOR4 = (0x200 + 7)

//out port 3
const ch_MOTORDIR = (0x300 + 15)
const ch_LIGHT_STOP = (0x300 + 14)
const ch_LIGHT_COMMAND1 = (0x300 + 13)
const ch_LIGHT_COMMAND2 = (0x300 + 12)
const ch_LIGHT_COMMAND3 = (0x300 + 11)
const ch_LIGHT_COMMAND4 = (0x300 + 10)
const ch_LIGHT_UP1 = (0x300 + 9)
const ch_LIGHT_UP2 = (0x300 + 8)

//out port 2
const ch_LIGHT_DOWN2 = (0x300 + 7)
const ch_LIGHT_UP3 = (0x300 + 6)
const ch_LIGHT_DOWN3 = (0x300 + 5)
const ch_LIGHT_DOWN4 = (0x300 + 4)
const ch_LIGHT_DOOR_OPEN = (0x300 + 3)
const ch_LIGHT_FLOOR_IND2 = (0x300 + 1)
const ch_LIGHT_FLOOR_IND1 = (0x300 + 0)

//out port 0
const ch_MOTOR = (0x100 + 0)

//non-existing ports (for alignment)
const ch_BUTTON_DOWN1 = -1
const ch_BUTTON_UP4 = -1
const ch_LIGHT_DOWN1 = -1
const ch_LIGHT_UP4 = -1
