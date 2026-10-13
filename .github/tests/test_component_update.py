import importlib.util
from pathlib import Path
import sys
import unittest
from unittest.mock import patch
sys.path.insert(0,str(Path('.github/scripts').resolve()))
import component_update as m
class Tests(unittest.TestCase):
 def test_stable_order_and_legacy_timestamp(self):
  self.assertGreater(m.version('v0.0.21'),m.version('v0.0.9'))
  self.assertIsNone(m.version('v0.0.2-20240624082446'))
  self.assertEqual(m.bump('v0.2.12'),'v0.2.13')
 def test_prerelease_major_and_legacy_are_not_auto_adopted(self):
  with patch.object(m,'rows',return_value=[{'name':'v0.0.21-rc1'},{'name':'v1.0.0'},{'name':'v0.0.20'}]):
   self.assertEqual(m.latest_component(m.PREFIX+'basics','v0.0.20'),'v0.0.20')
  with patch.object(m,'rows',return_value=[{'name':'v0.0.2-20240624082446'},{'name':'v0.0.2-20200101000000'}]):
   self.assertEqual(m.latest_component(m.PREFIX+'defaultset','v0.0.2-20240624082446'),'v0.0.2-20240624082446')
 def test_timestamp_updates_are_monotonic_and_calendar_valid(self):
  with patch.object(m,'rows',return_value=[{'name':'v0.0.2-20261013123456'},{'name':'v0.0.2-99999999999999'},{'name':'v0.0.2-20200101000000'}]):
   self.assertEqual(m.latest_component(m.PREFIX+'defaultset','v0.0.2-20240624082446'),'v0.0.2-20261013123456')
 def test_private_component_not_used_until_all_platforms_are_published(self):
  release={'tag_name':'v0.0.21','draft':True,'prerelease':False,'assets':[]}
  with patch.object(m,'rows',side_effect=[[{'name':'v0.0.21'}],[release]]):
   self.assertEqual(m.latest_component(m.PREFIX+'security','v0.0.20'),'v0.0.20')
 def test_failed_self_test_stops_before_commit_tag_or_dispatch(self):
  current={m.PREFIX+'basics':'v0.0.20'}
  calls=[]
  with patch.object(m,'requirements',side_effect=[current,{m.PREFIX+'basics':'v0.0.21'}]),patch.object(m,'application_version',return_value='v0.2.12'),patch.object(m,'latest_component',return_value='v0.0.21'),patch.object(m,'command',side_effect=lambda args,*a:calls.append(args)),patch.object(m,'sync_versions'),patch.object(m,'workspace'),patch.object(m,'test_consumer',side_effect=RuntimeError('fixture test failed')),patch.object(m,'recover_release') as recover:
   with self.assertRaises(RuntimeError):m.update('ecs')
   recover.assert_not_called()
  self.assertFalse(any(c[:2]==['git','commit'] or c[:2]==['git','push'] for c in calls))
if __name__=='__main__':unittest.main(verbosity=2)
