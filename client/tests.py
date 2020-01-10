import unittest
from Auth import User

class TestUserMethods(unittest.TestCase):

    def setUp(self):
        self.user = User(name="Joe Schmoe", is_admin=True, email="j.shmoe@example.org")

    def test_create(self):
        self.user.create()
        self.assertIsNotNone(self.user.user_id)

    def test_get(self):
        new_user = User(user_id=self.user.user_id)
        new_user.get()

        self.assertEqual(new_user.name, self.user.name)
        self.assertEqual(new_user.email, self.user.email)
        self.assertEqual(new_user.is_admin, self.user.is_admin)

    def test_delete(self):
        self.user.delete()

        deleted_user = User(user_id = self.user.user_id)

        deleted_user.get()

        self.assertIsNone(deleted_user.name)
        self.assertIsNone(deleted_user.email)

    def tearDown(self):
        pass



if __name__ == "__main__":
    unittest.main()
