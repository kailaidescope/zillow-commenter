console.log("User ID: ", getLocalUserId());

// Sets a unique user ID in localStorage if it doesn't exist
//
// Note: The localstorage persists between browser sessions, and incognito mode
function setUserId() {
    let userId = window.localStorage.getItem('zillow_commenter_user_id');
    if (!userId) {
        getNewUserId((result, error=null) => {
            if (error) {
                // If there's an error retrieving the user ID, log it
                console.error('Error retrieving user ID:', error);
                //setUserId();
                return;
            }
            if (result) {
                // If a user ID is found, parse it and log it
                const parsedResult = JSON.parse(result);
                console.log('Retrieved user ID:', parsedResult.user_id);
                window.localStorage.setItem('zillow_commenter_user_id', parsedResult.user_id);
            } else {
                // If no user ID is found, generate a new one
                console.log('No user ID found, generating a new one.');
                //setUserId();
            }
        });
        /* userId = "user_" + Math.random().toString(36).substring(2, 15);
        window.localStorage.setItem('zillow_commenter_user_id', userId); */
    }
}

setUserId();

// Retrieves the user ID from localStorage
function getLocalUserId() {
    return window.localStorage.getItem('zillow_commenter_user_id');
}

console.log(getLocalUserId());


// Tab switching logic
document.querySelectorAll('.tab').forEach(tab => {
    tab.addEventListener('click', function() {
        // Remove active from all tabs and contents
        document.querySelectorAll('.tab').forEach(t => t.classList.remove('active'));
        document.querySelectorAll('.tab-content').forEach(c => c.classList.remove('active'));
        // Add active to clicked tab and corresponding content
        this.classList.add('active');
        document.getElementById(this.dataset.tab).classList.add('active');
    });
});

// Sample comments array (just strings)
const sampleComments = [
    "Great property, loved the backyard!",
    "Needs some renovation, but has potential.",
    "Amazing location and spacious rooms.",
    "Too expensive for the area.",
    "Would love to schedule a tour."
];

// Sample comment objects
const sampleCommentObjects = [
    {
        datePosted: "2024-06-01",
        username: "user123",
        commentText: "Great property, loved the backyard!"
    },
    {
        datePosted: "2024-06-02",
        username: "houseHunter",
        commentText: "Needs some renovation, but has potential."
    },
    {
        datePosted: "2024-06-03",
        username: "realtyFan",
        commentText: "Amazing location and spacious rooms."
    },
    {
        datePosted: "2024-06-04",
        username: "skeptic42",
        commentText: "Too expensive for the area."
    },
    {
        datePosted: "2024-06-05",
        username: "tourSeeker",
        commentText: "Would love to schedule a tour."
    }
];

// Function to populate comments list in the DOM
async function populateComments() {
    // Get comments element from the DOM
    const commentsListElement = document.querySelector('.comments-list');
    console.log('Populating comments.');
    if (!commentsListElement) return;

    // Clear existing comments
    commentsListElement.innerHTML = '';

    const listingId = await getListingID();

    if (!listingId) {
        displayComments(null);
        console.error("No valid listing ID found in the current URL.");
        return;
    }

    // Fetch comments from the API using the listing ID
    getCommentsByListingId(listingId, displayComments);
}

// Function to display comments in the list
function displayComments(result, error=null) {
    // Get comments element from the DOM
    const commentsListElement = document.querySelector('.comments-list');
    if (!commentsListElement) return;
    
    // Clear existing comments
    commentsListElement.innerHTML = '';

    if (error) {
        console.error('Error fetching comments:', error);
        const li = document.createElement('li');
        li.innerHTML = '<strong>Error fetching comments.</strong> Please try again later.';
        commentsListElement.appendChild(li);
        const submitButton = document.querySelector('#comment-form button[type="submit"]');
        if (submitButton) {
            submitButton.disabled = true;
            submitButton.style.backgroundColor = '#ccc';
        }
        return;
    }

    let comments = null;

    //console.log('Fetched comments:', result);
    if (result) {
        try {
            comments = JSON.parse(result);
            //console.log('Parsed comments:', comments);
        } catch (error) {
            console.error('Error parsing comments:', error);
            const li = document.createElement('li');
            li.textContent = 'Error loading comments.';
            commentsListElement.appendChild(li);
            return;
        }
    } else {
        // If no comments are returned, check if we have a valid listing ID
        getListingID().then(listingId => {
            if (!listingId) {
                // If no valid listing ID, disable the submit button and show an error message
                console.error("No valid listing ID found in the current URL.");
                const li = document.createElement('li');
                li.textContent = 'Not on a commentable zillow listing page.';
                commentsListElement.appendChild(li);
                return;
            } else {
                // If we have a valid listing ID but no comments, display a message
                const li = document.createElement('li');
                li.textContent = 'No comments found for this listing.';
                commentsListElement.appendChild(li);
            }
        });
        
        return;
    }

    // Check if there are any comments
    if (comments !== null) {
        console.log('Displaying comments: ', comments);
    } else {
        console.log('Displaying comments: comments is null');
    }

    if (!Array.isArray(comments)) {
        console.error('Invalid comments data:', comments);
    }

    if (!comments || comments.length === 0) {
        // If no comments, display a message
        const li = document.createElement('li');
        li.textContent = 'No comments available for this listing.';
        commentsListElement.appendChild(li);
        return;
    }

    // Populate the comments list
    comments.forEach(comment => {
        const li = document.createElement('li');
        // Convert Unix second timestamp to readable date or time
        let dateStr = 'Unknown date';
        console.log('Comment timestamp:', comment.timestamp);
        console.log("of type", typeof comment.timestamp);
        console.log('Comment:', comment.timestamp !== undefined && comment.timestamp !== null && !isNaN(Number(comment.timestamp)));
        // Check if timestamp exists and is a valid int64 in seconds
        if (comment.timestamp !== undefined && comment.timestamp !== null && !isNaN(Number(comment.timestamp))) {
            console.log("Raw timestamp:", comment.timestamp);

            // Convert int64 microseconds to milliseconds for JS Date
            const dateObj = new Date(comment.timestamp / 1000);
            console.log("Converted dateObj:", dateObj);

            const now = new Date();
            console.log("Current date (now):", now);

            const diffMs = now - dateObj;
            console.log("Difference in ms (now - dateObj):", diffMs);

            const diffDays = Math.floor(diffMs / (1000 * 60 * 60 * 24));
            console.log("Difference in days:", diffDays);

            const isToday = dateObj.getFullYear() === now.getFullYear() &&
                            dateObj.getMonth() === now.getMonth() &&
                            dateObj.getDate() === now.getDate();
            console.log("Is today:", isToday);

            if (isToday) {
                dateStr = dateObj.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' });
                console.log("Formatted as time (today):", dateStr);
            } else if (diffDays === 1) {
                dateStr = "1 day ago";
                console.log("Formatted as 1 day ago");
            } else if (diffDays > 1 && diffDays < 7) {
                dateStr = `${diffDays} days ago`;
                console.log(`Formatted as ${diffDays} days ago`);
            } else if (diffDays >= 7) {
                dateStr = dateObj.toLocaleDateString();
                console.log("Formatted as date string:", dateStr);
            }
        }
        li.innerHTML = `<strong>${comment.username}</strong> <span style="font-size: 0.85em; color: #555;">${dateStr}</span><br>${comment.comment_text}`;
        commentsListElement.appendChild(li);
    });
}

// Call this function to populate comments when the popup is opened
populateComments();

// Function to show the current URL in the comments tab
function displayURL() {
    const commentsTabContent = document.getElementById('comments');
    if (commentsTabContent) {
        const urlHeader = document.createElement('div');
        urlHeader.className = 'comments-url-header';
        
        // Get the current tab's URL using the Chrome extension API
        chrome.tabs.query({ currentWindow: true, active: true }, function (tabs) {
            console.log(tabs[0].url);
            if (tabs[0] && tabs[0].url) {
                urlHeader.textContent = `Comments for: ${tabs[0].url}`;
            } else {
                urlHeader.textContent = 'Comments for: [unknown URL]';
            }
        });
        // Insert header at the top of the comments tab content
        commentsTabContent.insertBefore(urlHeader, commentsTabContent.firstChild);
    }
}

// Call this function to display the current URL in the comments tab
//displayURL()


// Handles the comment form submission
document.getElementById('comment-form').addEventListener('submit', handleCommentSubmission);

// Compiles the form data and user data into a struct for submission
async function handleCommentSubmission(event) {
    // Stop default form submission behavior
    event.preventDefault();

    // Disable the form for 5 seconds to prevent multiple submissions
    const submitButton = event.target.querySelector('button[type="submit"]');
    submitButton.disabled = true;
    submitButton.style.backgroundColor = '#ccc';
    setTimeout(() => {
        submitButton.disabled = false;
        submitButton.style.backgroundColor = '';
    }, 3000);

    // Get comment data
    const username = document.getElementById('username-input').value.trim();
    const commentText = document.getElementById('comment-input').value.trim();
    const listingId = await getListingID();
    if (!listingId) {
        console.error("No valid listing ID found in the current URL.");
        return;
    }

    // Compile the comment object
    const commentObj = {
        userId: getLocalUserId(),
        listingId: listingId,
        username: username,
        commentText: commentText,
    };

    //console.log('Form submission:', commentObj);

    // Display and post the comment
    //displaySubmittedComment(commentObj)
    postComment(commentObj, (result, error) => {
        // Log the result or error
        console.log('Comment posted:', result, error);

        // Display the updated comments list after posting
        getCommentsByListingId(listingId, displayComments)
    });

    // Clear the form fields after submission
    document.getElementById('username-input').value = '';
    document.getElementById('comment-input').value = '';
}

// Gets the listing ID from the current URL
// zillow's URL format is "https://www.zillow.com/homedetails/listing-street-name/1234567890_zpid/", 
// from which you would extract "1234567890"
async function getListingID() {
    let listingURL = '';
    // Get the current tab's URL using the Chrome extension API
    const tabs = await chrome.tabs.query({ currentWindow: true, active: true });
    //console.log("Current tabs:", tabs);
    if (tabs.length > 0 && tabs[0].url) {
        listingURL = tabs[0].url;
    }

    // Extract the listing ID from the URL

    // Find section of the URL that ends with "_zpid"
    const urlParts = listingURL.split('/');
    const zpidIndex = urlParts.findIndex(part => part.endsWith('_zpid'));
    
    if (zpidIndex !== -1 && urlParts[zpidIndex]) {
        // Gets the listing ID by removing the "_zpid" suffix
        const listingID = urlParts[zpidIndex].replace('_zpid', '');
        //console.log("Listing ID found:", listingID);
        return listingID;
    }
    // If no valid listing ID is found, return null
    // Disable the submit button and show an error message
    const submitButton = document.querySelector('#comment-form button[type="submit"]');
    if (submitButton) {
        submitButton.disabled = true;
        submitButton.style.backgroundColor = '#ccc';
    }
    console.error("No valid listing ID found in the current URL.");
    return null; // No valid listing ID found
}

// Displays the submitted comment in the extension popup
function displaySubmittedComment(commentObj) {
    const form = document.getElementById('comment-form');
    if (!form) return;

    let submittedSection = document.getElementById('submitted-comment');
    if (!submittedSection) {
        submittedSection = document.createElement('div');
        submittedSection.id = 'submitted-comment';
        form.parentNode.insertBefore(submittedSection, form.nextSibling);
    }

    submittedSection.innerHTML = `
        <h4>Submitted Comment</h4>
        <div><strong>Username:</strong> ${commentObj.username}</div>
        <div><strong>Comment:</strong> ${commentObj.commentText}</div>
        <div><strong>Page:</strong> ${commentObj.listingId}</div>
        <div><strong>User ID:</strong> ${commentObj.userId}</div>
    `;
}

// Fetches the list of comments for a specific listing from the API
function getCommentsByListingId(listingId, callbackFunc) {
    if (!listingId) {
        console.error("No valid listing ID provided.");
        return [];
    }

    var requestOptions = {
    method: 'GET',
    redirect: 'follow'
    };

    fetch("http://localhost:3000/api/v1/comments/"+listingId, requestOptions)
        .then(response => response.text())
        .then(result => callbackFunc(result))
        .catch(error => callbackFunc(null, error));
}

async function postComment(commentObj, callbackFunc) {

    // Collect comment data
    let listingId = await getListingID();
    if (!listingId) {
        console.error("No valid listing ID found in the current URL.");
        return;
    }

    // Prepare form data for API
    var myHeaders = new Headers();
    myHeaders.append("Content-Type", "application/x-www-form-urlencoded");

    var urlencoded = new URLSearchParams();
    urlencoded.append("listing_id", listingId);
    urlencoded.append("user_id", getLocalUserId());
    urlencoded.append("username", commentObj.username);
    urlencoded.append("comment_text", commentObj.commentText);

    var requestOptions = {
    method: 'POST',
    headers: myHeaders,
    body: urlencoded,
    redirect: 'follow'
    };

    // Send POST request to the API
    fetch("http://localhost:3000/api/v1/comments", requestOptions)
        .then(response => response.text())
        .then(result => callbackFunc(result))
        .catch(error => callbackFunc(null, error));
}

// getNewUserId retrieves a new V7 (Time-based) UUID from the API
function getNewUserId(callbackFunc) {
    var requestOptions = {
        method: 'GET',
        redirect: 'follow'
    };

    fetch("http://localhost:3000/api/v1/user/user_id", requestOptions)
        .then(response => response.text())
        .then(result => callbackFunc(result))
        .catch(error => callbackFunc(null, error));
}
