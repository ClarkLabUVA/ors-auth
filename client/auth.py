import requests
import json
import logging
import urllib3

# temporary to prevent warnings for self signed cert for local development
urllib3.disable_warnings(urllib3.exceptions.InsecureRequestWarning)

AUTH_SERVER = "https://localhost:8080/"
VERIFY = False

class Error():
    def __init__(self, **kwargs):
        self.response = kwargs.get("response")
        self.message = kwargs.get("message")
        self.status_code = kwargs.get("status_code")

class User():
    def __init__(self, **kwargs):
        self.user_id = kwargs.get("user_id")
        self.name = kwargs.get("name")
        self.email = kwargs.get("email")
        self.is_admin = kwargs.get("is_admin", False)
        self.access_token = kwargs.get("access_token")
        self.refresh_token = kwargs.get("refresh_token")

    def get(self):
        ''' Get this user's record

        Requires the user_id or the user_email to be present
        '''

        if self.user_id is None:
            return "User ID cannot be None"

        response = requests.get(AUTH_SERVER + "user/" + self.user_id, verify=VERIFY)

        if response.status_code != 200:
            return response

        # FIXME: load response data into class
        response_body = json.loads(response.content.decode("utf-8"))

        self.name = response_body.get("name")
        self.email = response_body.get("email")
        self.is_admin = response_body.get("is_admin")

    def create(self):
        ''' Create a new user record
        '''

        response = requests.post(
            url=AUTH_SERVER + "user/",
            data=json.dumps({
                "@type": "Person",
                "name": self.name,
                "email": self.email,
                "is_admin": self.is_admin
            }),
            verify=VERIFY
        )

        if response.status_code != 201:
            return response

        # FIXME: load response data into class ('@id' element into self.user_id)
        response_body = json.loads(response.content.decode("utf-8"))

        self.id = response_body.get("@id")

        if self.id is None:
            return response

    def delete(self):
        response = requests.delete(
            url=AUTH_SERVER + "user/" + self.id,
            verify = VERIFY
        )

        if response.status_code != 200:
            return response

        return response

    def logout(self):
        pass

    def refresh_token(self):
        ''' Use the Refresh Token to Grant a New Access Token
        '''
        pass

    def list_resources(self):
        ''' List all resources this user owns
        '''
        pass

    def list_groups(self):
        ''' List all groups this user is a member of
        '''
        pass

    def list_policies(self):
        ''' List all policies effecting this user
        '''
        pass

    def list_challenges(self):
        ''' List all challenges this user has made
        '''
        pass

if __name__== '__main__':
    print("CLI Goes Here")
