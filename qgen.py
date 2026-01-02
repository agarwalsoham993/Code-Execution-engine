import random 
import json
import os

q = 2

# Define the logic for the problem
def function(input_val):
    if input_val > 0:
        return 2 * input_val + 1
    else:
        return 2 * input_val - 1
  
# Create directory if it doesn't exist
output_dir = f"./Questions/{q}"
os.makedirs(output_dir, exist_ok=True)

# Generate Question Text
# with open(f"{output_dir}/question.txt", "w") as f:
#    f.write("Double and Add/Sub\nGiven an integer x, return 2x+1 if x>0, else 2x-1.")

test_cases = []

for i in range(1, 101):
    input_val = random.randint(-100, 100)
    expected_output = function(input_val)
    
    # JSON structure required by worker
    test_cases.append({
        "id": str(i),
        "input": str(input_val),
        "expected_output": str(expected_output)
    })

# Write the single JSON file
output_path = os.path.join(output_dir, "tests.json")
with open(output_path, 'w') as f:
    json.dump(test_cases, f, indent=4)

print(f"Generated {len(test_cases)} test cases at {output_path}")
