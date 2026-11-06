from pprint import pprint

class Game:
    def __init__(self):
            # self.numbers = tuple("123456789")
            # self.letters = tuple("abcdefghi")
            self.board = []
            # declare hexagon board
            smol = 4
            biig = 9
            for x in range(9):
                self.board.append({})

                for y in range(smol, biig):
                    self.board[x][y] = None

                # Shift number bounds
                if smol > 0:
                    smol -= 1
                else:
                    biig -= 1

    def initBoard(self, newBoard=None):
        if newBoard != None:
            self.board = newBoard
            return
        newBoard = self.board
        



    def coordPrint(self, x: int, y: int):
        print(f"({"abcdefghi"[x]}, {y+1})")

    def posExists(self, x: int, y: int) -> bool:
        if x in range(9):
             if y in self.board[x]:
                    return True
        return False

    # letter (x) and number (y) position, ang direction starting at 0 right and rotating counterclockwise
    def getSpacesInRow(self, x: int, y: int, ang: int):
        if not self.posExists(x,y):
            raise Exception(f"Invalid position! ({x}, {y}) doesn't exist on board.")
        
        output = []
        if ang == 0:
            iterator = zip(range(x+1, 9), range(y-1, -1, -1))
        elif ang == 1:
            iterator = zip(range(x+1, 9), (y,)*9)
        elif ang == 2:
            iterator = zip((x,)*9, range(y+1, 9))
        elif ang == 3:
            iterator = zip(range(x-1, -1, -1), range(y+1, 9))
        elif ang == 4:
            iterator = zip(range(x-1, -1, -1), (y,)*9)
        elif ang == 5:
            iterator = zip((x,)*9, range(y-1, -1, -1))
        else:
            raise Exception(f"Invalid angle of direction: {ang}. Must be between 0 and 5.")
        
        for i, j in iterator:
            if self.posExists(i,j):
                output.append({'x': i, 'y': j})

        return output
    
myGame = Game()
#pprint(myGame.board)
while True:
    x = int(input("x: "))
    y = int(input("y: "))
    ang = int(input("dir: "))
    result = myGame.getSpacesInRow(x, y, ang)
    for coord in result:
        myGame.coordPrint(coord['x'], coord['y'])