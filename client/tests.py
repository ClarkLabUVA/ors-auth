import unittest
from Auth import User

class TestUserMethods(unittest.TestCase):

    def setUp(self):
        self.user = User(name="Joe Schmoe", is_admin=True, email="j.shmoe@example.org")
        self.user.delete()

    def test_crud(self):

        self.user.create()
        self.assertIsNotNone(self.user.user_id)

        # test getting the user by Id
        new_user = User(user_id=self.user.user_id)
        new_user.get()

        self.assertEqual(new_user.name, self.user.name)
        self.assertEqual(new_user.email, self.user.email)
        self.assertEqual(new_user.is_admin, self.user.is_admin)

        # test deletion
        self.user.delete()

        # prove the record is deleted
        deleted_user = User(user_id = self.user.user_id)
        deleted_user.get()

        self.assertIsNone(deleted_user.name)
        self.assertIsNone(deleted_user.email)

    def tearDown(self):
        self.user.delete()

if __name__ == "__main__":
    unittest.main()
