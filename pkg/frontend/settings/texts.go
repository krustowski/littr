package settings

const (
	AlertUserDeletion = "Please note that this action is irreversible!"

	InfoAboutYouTextarea = "This textarea is to hold your current status, a brief info about you, or just anything up to 100 characters."
	InfoWebsiteLink      = "Down below, you can enter a link to your personal homepage. The link will then be visible to others via the user modal on the users (flowers) page."
	InfoGravatarLinking  = "Your avatar can be linked to your e-mail address. In such case, your e-mail address needs to be registered with the #link class='bold' to='http://gravatar.com/profile/avatars'#Gravatar.com##link#. Note: if you just changed your icon at Gravatar.com, and the thumbnail above shows the old avatar, some intercepting cache probably has the resource cached (e.g. your browser). You may need to wait for some time for the change to propagate through the network."
	InfoUIMode           = "#bold class='blue-text'#The UI mode##bold# can be adjusted according to the user's choice. The mode may differ on the other sessions when logged-in on multiple devices."
	InfoLocalTimeMode    = "#bold class='blue-text'#The local time mode##bold# is a feature allowing you to see any post's (or poll's) timestamp according to your device's setting (mainly the timezone). When disabled, the server time is used instead."
	InfoLiveMode         = "#bold class='blue-text'#The live mode##bold# is a feature for the live flow experience. When enabled, a notice about some followed account's/user's new post is shown on the bottom of the page."
	InfoPrivateMode      = "#bold class='blue-text'#Private account##bold# is a feature allowing one to be hidden on the site. When enabled, other accounts/users need to ask you to follow you (the follow request will show on the users page). Any reply to your post will be shown as redacted (a private content notice) to those not following you."
	InfoNotifications    = "#bold class='blue-text'#Reply##bold# notifications are fired when someone posts a reply to your post. #break###break##bold class='blue-text'#Mention##bold# notifications are fired when someone mentions you via the at-sign (@) handler in their post (e.g. Hello, @example!).#break###break# #break###break#You will be prompted for the notification permission, which is required if you want to subscribe to the notification service. Your device's UUID (unique identification string) will be saved in the database to be used by the notification service. You can delete any subscribed device any time (if listed below)."
)
