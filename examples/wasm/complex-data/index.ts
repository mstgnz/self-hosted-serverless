export function execute(input: string): string {
  const data = JSON.parse(input);
  
  // Extract user data
  const user = data.user || {};
  const name = user.name || "Unknown";
  const age = user.age || 0;
  
  // Extract items
  const items = data.items || [];
  
  // Process data
  const result = {
    greeting: `Hello, ${name}!`,
    age_in_months: age * 12,
    item_count: items.length,
    processed_items: items.map((item: string) => `Processed: ${item}`)
  };
  
  return JSON.stringify(result);
} 