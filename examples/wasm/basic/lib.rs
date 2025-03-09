#[no_mangle]
pub extern "C" fn execute(input_ptr: i32, input_len: i32) -> i32 {
    // Parse input JSON
    let input_slice = unsafe {
        std::slice::from_raw_parts(input_ptr as *const u8, input_len as usize)
    };
    
    let input: serde_json::Value = match serde_json::from_slice(input_slice) {
        Ok(v) => v,
        Err(_) => return -1,
    };
    
    // Extract name from input
    let name = input["name"].as_str().unwrap_or("World");
    
    // Create response
    let response = format!("{{\"message\":\"Hello, {} from WebAssembly!\"}}", name);
    
    // Allocate memory for response
    let response_bytes = response.into_bytes();
    let response_ptr = Box::into_raw(response_bytes.into_boxed_slice()) as i32;
    
    // Return pointer to response
    response_ptr
} 