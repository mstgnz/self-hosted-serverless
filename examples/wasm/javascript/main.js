export function execute(input) {
  const data = JSON.parse(input);
  
  // Process data
  const result = {
    message: `Hello from JavaScript in WebAssembly!`,
    input: data,
    timestamp: new Date().toISOString()
  };
  
  return JSON.stringify(result);
} 