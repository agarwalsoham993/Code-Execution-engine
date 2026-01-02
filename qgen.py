import random 
import os

q = 2

def function(input):
  if(input>0):
    return 2*input+1
  else:
    return 2*input-1
  
os.mkdir(f"./Questions/{q}")
os.mkdir(f"./Questions/{q}/input")
os.mkdir(f"./Questions/{q}/output")

for i in range(1,101):
    input = random.randint(-100,100)
    with open(f"./Questions/{q}/input/input_{i}.txt",'w') as f:
      f.write(str(input))
    with open(f"./Questions/{q}/output/output_{i}.txt",'w') as f:
      f.write(str(function(input)))
