const url = "http://localhost:8080/api/sdsad";

async function makeRequest() {
  const startTime = performance.now(); // Capture the start time

  try {
    const response = await fetch(url, {
      credentials: "omit",
      method: "POST",
      body: JSON.stringify({ hello: "dsad" }), // Serialize the object to a JSON string
      headers: {
        "Content-Type": "application/json", // Don't forget to set the content type to JSON
      },
      mode: "cors",
    });

    if (!response.ok) {
      throw new Error(`HTTP error! Status: ${response.status}`);
    }

    const data = await response.json(); // Assuming the response is JSON
    console.log("Response data:", data);
  } catch (error) {
    console.error("Error:", error);
  } finally {
    const endTime = performance.now(); // Capture the end time
    const duration = endTime - startTime; // Calculate the time difference
    console.log(`Request took ${duration.toFixed(2)} ms`); // Print the time in milliseconds
  }
}

// Call the function to make the request
makeRequest();
